package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/updates"
	"github.com/gotd/td/tg"
)

// Config holds the daemon configuration
type Config struct {
	AppID                 int    `yaml:"app_id" env:"APP_ID"`
	AppHash               string `yaml:"app_hash" env:"APP_HASH"`
	SessionFile           string `yaml:"session_file" env:"SESSION_FILE"`
	Phone                 string `yaml:"phone" env:"PHONE"`
	Password              string `yaml:"password" env:"PASSWORD"`
	ResponseMsg           string `yaml:"response_message" env:"RESPONSE_MSG"`
	ResponseTimeoutHours  int    `yaml:"response_timeout_hours" env:"RESPONSE_TIMEOUT_HOURS"`
	LogLevel              string `yaml:"log_level" env:"LOG_LEVEL"`
	LogFile               string `yaml:"log_file" env:"LOG_FILE"`
	EnableDaemonMode      bool   `yaml:"enable_daemon_mode" env:"ENABLE_DAEMON_MODE"`
}

// DefaultConfig returns a config with default values
func DefaultConfig() *Config {
	return &Config{
		SessionFile:          "session.json",
		ResponseMsg:          "Hi! I'm no longer using Telegram. Please contact me via email or other means.",
		ResponseTimeoutHours: 24,
		LogLevel:             "info",
		LogFile:              "",
		EnableDaemonMode:     false,
	}
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configPath string) (*Config, error) {
	config := DefaultConfig()

	// Load from YAML file if it exists
	if configPath != "" {
		if _, err := os.Stat(configPath); err == nil {
			data, err := os.ReadFile(configPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read config file: %w", err)
			}

			if err := yaml.Unmarshal(data, config); err != nil {
				return nil, fmt.Errorf("failed to parse config file: %w", err)
			}
		}
	}

	// Override with environment variables
	if val := os.Getenv("APP_ID"); val != "" {
		if _, err := fmt.Sscanf(val, "%d", &config.AppID); err != nil {
			return nil, fmt.Errorf("invalid APP_ID: %w", err)
		}
	}
	if val := os.Getenv("APP_HASH"); val != "" {
		config.AppHash = val
	}
	if val := os.Getenv("SESSION_FILE"); val != "" {
		config.SessionFile = val
	}
	if val := os.Getenv("PHONE"); val != "" {
		config.Phone = val
	}
	if val := os.Getenv("PASSWORD"); val != "" {
		config.Password = val
	}
	if val := os.Getenv("RESPONSE_MSG"); val != "" {
		config.ResponseMsg = val
	}
	if val := os.Getenv("RESPONSE_TIMEOUT_HOURS"); val != "" {
		if _, err := fmt.Sscanf(val, "%d", &config.ResponseTimeoutHours); err != nil {
			return nil, fmt.Errorf("invalid RESPONSE_TIMEOUT_HOURS: %w", err)
		}
	}
	if val := os.Getenv("LOG_LEVEL"); val != "" {
		config.LogLevel = val
	}
	if val := os.Getenv("LOG_FILE"); val != "" {
		config.LogFile = val
	}
	if val := os.Getenv("ENABLE_DAEMON_MODE"); val != "" {
		config.EnableDaemonMode = val == "true" || val == "1"
	}

	return config, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.AppID == 0 {
		return fmt.Errorf("app_id is required")
	}
	if c.AppHash == "" {
		return fmt.Errorf("app_hash is required")
	}
	if c.Phone == "" {
		return fmt.Errorf("phone is required")
	}
	if c.ResponseTimeoutHours < 1 {
		return fmt.Errorf("response_timeout_hours must be at least 1")
	}
	return nil
}

// TelegramDaemon represents the main daemon structure
type TelegramDaemon struct {
	config          *Config
	client          *telegram.Client
	contacts        map[int64]bool      // Cache of contact user IDs
	respondedUsers  map[int64]time.Time // Track users who have been responded to
	responseMutex   sync.RWMutex        // Mutex for thread-safe access to respondedUsers
	responseTimeout time.Duration       // How long to wait before responding to the same user again
}

// NewTelegramDaemon creates a new daemon instance
func NewTelegramDaemon(config *Config) *TelegramDaemon {
	return &TelegramDaemon{
		config:          config,
		contacts:        make(map[int64]bool),
		respondedUsers:  make(map[int64]time.Time),
		responseTimeout: time.Duration(config.ResponseTimeoutHours) * time.Hour,
	}
}

// authenticator handles the authentication flow
type authenticator struct {
	phone    string
	password string
}

func (a *authenticator) Phone(ctx context.Context) (string, error) {
	return a.phone, nil
}

func (a *authenticator) Password(ctx context.Context) (string, error) {
	return a.password, nil
}

func (a *authenticator) Code(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
	fmt.Print("Enter the code sent to your phone: ")
	var code string
	if _, err := fmt.Scanln(&code); err != nil {
		return "", err
	}
	return code, nil
}

func (a *authenticator) AcceptTermsOfService(ctx context.Context, tos tg.HelpTermsOfService) error {
	// Automatically accept terms of service if prompted
	log.Printf("Accepting Telegram Terms of Service: %s", tos.Text)
	return nil
}

func (a *authenticator) SignUp(ctx context.Context) (auth.UserInfo, error) {
	// This daemon does not support signing up new users.
	// If this is called, it means the phone number is not registered with Telegram.
	return auth.UserInfo{}, fmt.Errorf("sign up not supported")
}

// loadContacts fetches and caches the user's contact list
func (d *TelegramDaemon) loadContacts(ctx context.Context) error {
	api := d.client.API()
	
	// Get contacts
	contactsResult, err := api.ContactsGetContacts(ctx, 0) // Pass 0 directly as hash
	if err != nil {
		return fmt.Errorf("failed to get contacts: %w", err)
	}

	// Process contacts based on the result type
	switch contacts := contactsResult.(type) {
	case *tg.ContactsContacts:
		log.Printf("Loaded %d contacts", len(contacts.Contacts))
		
		// Clear existing contacts cache
		d.contacts = make(map[int64]bool)
		
		// Add contacts to cache
		for _, contact := range contacts.Contacts {
			d.contacts[contact.UserID] = true
		}
		
		// Log contact user IDs for debugging
		if d.config.LogLevel == "debug" {
			log.Printf("Contact user IDs: %v", d.getContactIDs())
		}
		
	case *tg.ContactsContactsNotModified:
		log.Println("Contacts not modified, using cached version")
		
	default:
		return fmt.Errorf("unexpected contacts result type: %T", contacts)
	}

	return nil
}

// getContactIDs returns a slice of contact user IDs for debugging
func (d *TelegramDaemon) getContactIDs() []int64 {
	var ids []int64
	for id := range d.contacts {
		ids = append(ids, id)
	}
	return ids
}

// isContact checks if a user ID is in the contacts list
func (d *TelegramDaemon) isContact(userID int64) bool {
	return d.contacts[userID]
}

// shouldRespond checks if we should respond to a user (rate limiting)
func (d *TelegramDaemon) shouldRespond(userID int64) bool {
	d.responseMutex.RLock()
	lastResponse, exists := d.respondedUsers[userID]
	d.responseMutex.RUnlock()
	
	if !exists {
		return true
	}
	
	return time.Since(lastResponse) > d.responseTimeout
}

// markUserResponded marks a user as having been responded to
func (d *TelegramDaemon) markUserResponded(userID int64) {
	d.responseMutex.Lock()
	d.respondedUsers[userID] = time.Now()
	d.responseMutex.Unlock()
}

// sendResponse sends the predefined response message to a user
func (d *TelegramDaemon) sendResponse(ctx context.Context, userID int64, userName string) error {
	// Create message sender
	sender := message.NewSender(d.client.API())
	
	// Create input peer for the user
	inputPeer := &tg.InputPeerUser{
		UserID: userID,
	}
	
	// Send the response message
	_, err := sender.To(inputPeer).Text(ctx, d.config.ResponseMsg)
	if err != nil {
		return fmt.Errorf("failed to send response to %s (ID: %d): %w", userName, userID, err)
	}
	
	log.Printf("Sent response to %s (ID: %d): %s", userName, userID, d.config.ResponseMsg)
	return nil
}

// Start initializes and starts the daemon
func (d *TelegramDaemon) Start(ctx context.Context) error {
	// Create update dispatcher
	dispatcher := tg.NewUpdateDispatcher()
	
	// Create gaps handler for updates
	gaps := updates.New(updates.Config{
		Handler: dispatcher,
	})

	// Create telegram client
	d.client = telegram.NewClient(d.config.AppID, d.config.AppHash, telegram.Options{
		SessionStorage: &telegram.FileSessionStorage{
			Path: d.config.SessionFile,
		},
		UpdateHandler: gaps,
	})

	// Setup authentication flow
	flow := auth.NewFlow(
		&authenticator{
			phone:    d.config.Phone,
			password: d.config.Password,
		},
		auth.SendCodeOptions{},
	)

	// Setup message handlers
	d.setupMessageHandlers(&dispatcher)

	// Run the client
	return d.client.Run(ctx, func(ctx context.Context) error {
		// Perform authentication if necessary
		if err := d.client.Auth().IfNecessary(ctx, flow); err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}

		// Get self info
		self, err := d.client.Self(ctx)
		if err != nil {
			return fmt.Errorf("failed to get self info: %w", err)
		}

		log.Printf("Authenticated as: %s %s (ID: %d)", self.FirstName, self.LastName, self.ID)

		// Load contacts
		if err := d.loadContacts(ctx); err != nil {
			return fmt.Errorf("failed to load contacts: %w", err)
		}

		// Start gaps handler
		return gaps.Run(ctx, d.client.API(), self.ID, updates.AuthOptions{
			OnStart: func(ctx context.Context) {
				log.Println("Telegram daemon started successfully")
			},
		})
	})
}

// setupMessageHandlers configures the message handlers
func (d *TelegramDaemon) setupMessageHandlers(dispatcher *tg.UpdateDispatcher) {
	// Handle new private messages
	dispatcher.OnNewMessage(func(ctx context.Context, e tg.Entities, update *tg.UpdateNewMessage) error {
		return d.handleNewMessage(ctx, e, update)
	})

	// Handle new channel messages (optional, for debugging)
	if d.config.LogLevel == "debug" {
		dispatcher.OnNewChannelMessage(func(ctx context.Context, e tg.Entities, update *tg.UpdateNewChannelMessage) error {
			log.Printf("Channel message received: %+v", update.Message)
			return nil
		})
	}
}

// handleNewMessage processes incoming private messages
func (d *TelegramDaemon) handleNewMessage(ctx context.Context, e tg.Entities, update *tg.UpdateNewMessage) error {
	message, ok := update.Message.(*tg.Message)
	if !ok {
		if d.config.LogLevel == "debug" {
			log.Printf("Received non-message update: %T", update.Message)
		}
		return nil
	}

	// Skip outgoing messages (messages sent by us)
	if message.Out {
		if d.config.LogLevel == "debug" {
			log.Printf("Skipping outgoing message")
		}
		return nil
	}

	// Extract sender information
	var senderID int64
	var senderName string

	switch peer := message.PeerID.(type) {
	case *tg.PeerUser:
		senderID = peer.UserID
		
		// Get user info from entities
		if user, exists := e.Users[peer.UserID]; exists {
			senderName = fmt.Sprintf("%s %s", user.FirstName, user.LastName)
		}
		
	case *tg.PeerChat:
		if d.config.LogLevel == "debug" {
			log.Printf("Ignoring group chat message from chat ID: %d", peer.ChatID)
		}
		return nil
		
	case *tg.PeerChannel:
		if d.config.LogLevel == "debug" {
			log.Printf("Ignoring channel message from channel ID: %d", peer.ChannelID)
		}
		return nil
		
	default:
		if d.config.LogLevel == "debug" {
			log.Printf("Unknown peer type: %T", peer)
		}
		return nil
	}

	// Log the message details
	log.Printf("Message from %s (ID: %d): %s", senderName, senderID, message.Message)

	// Check if sender is a contact
	if d.isContact(senderID) {
		log.Printf("Message from contact %s (ID: %d)", senderName, senderID)
		
		// Check if we should respond (rate limiting)
		if d.shouldRespond(senderID) {
			log.Printf("Sending response to %s (ID: %d)", senderName, senderID)
			
			// Send the response
			if err := d.sendResponse(ctx, senderID, senderName); err != nil {
				log.Printf("Error sending response: %v", err)
				return err
			}
			
			// Mark user as responded to
			d.markUserResponded(senderID)
		} else {
			log.Printf("Already responded to %s (ID: %d) recently, skipping", senderID)
		}
	} else {
		log.Printf("Message from non-contact %s (ID: %d) - ignoring", senderID)
	}

	return nil
}

// setupLogging configures logging based on the configuration
func setupLogging(config *Config) error {
	if config.LogFile != "" {
		// Create log directory if it doesn't exist
		logDir := filepath.Dir(config.LogFile)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}

		// Open log file
		logFile, err := os.OpenFile(config.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}

		// Set log output to file
		log.SetOutput(logFile)
	}

	// Set log flags
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	return nil
}

func main() {
	// Determine config file path
	configPath := "config.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	// Load configuration
	config, err := LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Setup logging
	if err := setupLogging(config); err != nil {
		log.Fatalf("Failed to setup logging: %v", err)
	}

	log.Printf("Starting Telegram daemon with config from: %s", configPath)

	// Create daemon
	daemon := NewTelegramDaemon(config)

	// Setup signal handling for graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Start the daemon
	log.Println("Starting Telegram daemon...")
	if err := daemon.Start(ctx); err != nil {
		log.Fatalf("Daemon failed: %v", err)
	}

	log.Println("Telegram daemon stopped")
}


