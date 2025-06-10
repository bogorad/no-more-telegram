# Telegram Daemon

A Go daemon that connects to Telegram using the gotd library, listens for messages, ignores messages from non-contacts, and responds to contacts with a predefined message.

## Features

- ✅ Telegram authentication and connection
- ✅ Contact detection and filtering
- ✅ Auto-response to contacts only
- ✅ Rate limiting (configurable timeout between responses)
- ✅ YAML configuration file support
- ✅ Environment variable support
- ✅ Docker support
- ✅ Systemd service support
- ✅ Configurable logging

## Configuration

The daemon can be configured using a YAML file, environment variables, or a combination of both. Environment variables take precedence over the configuration file.

### Configuration File

Copy `config.yaml.example` to `config.yaml` and modify as needed:

```yaml
# Telegram API credentials (required)
app_id: 0  # Your app ID from https://my.telegram.org/apps
app_hash: ""  # Your app hash from https://my.telegram.org/apps

# Authentication (required)
phone: ""  # Your phone number in international format (e.g., +1234567890)
password: ""  # Your 2FA password (leave empty if 2FA is not enabled)

# Session storage
session_file: "session.json"  # Path to session file

# Response configuration
response_message: "Hi! I'm no longer using Telegram. Please contact me via email or other means."
response_timeout_hours: 24  # Hours to wait before responding to the same contact again

# Logging
log_level: "info"  # debug, info, warn, error
log_file: ""  # Leave empty to log to stdout, or specify a file path

# Daemon settings
enable_daemon_mode: false  # Run as a background daemon (Linux only)
```

### Environment Variables

- `APP_ID`: Your Telegram app ID (required)
- `APP_HASH`: Your Telegram app hash (required)  
- `PHONE`: Your phone number in international format (required)
- `PASSWORD`: Your 2FA password (optional, only if 2FA is enabled)
- `SESSION_FILE`: Path to session file (optional, defaults to "session.json")
- `RESPONSE_MSG`: Message to send to contacts (optional, has a default message)
- `RESPONSE_TIMEOUT_HOURS`: Hours between responses to same contact (optional, defaults to 24)
- `LOG_LEVEL`: Logging level (optional, defaults to "info")
- `LOG_FILE`: Log file path (optional, logs to stdout if not set)
- `ENABLE_DAEMON_MODE`: Run as daemon (optional, defaults to false)

## Setup

### Prerequisites

1. Get your APP_ID and APP_HASH from https://my.telegram.org/apps
2. Go 1.18 or later

### Installation

#### Option 1: Direct Installation

```bash
# Clone or download the source code
git clone <repository-url>
cd telegram-daemon

# Build the binary
go build -o telegram-daemon

# Copy and configure
cp config.yaml.example config.yaml
nano config.yaml

# Run
./telegram-daemon
```

#### Option 2: System Service (Linux)

```bash
# Build the binary
go build -o telegram-daemon

# Install as system service
sudo ./install.sh

# Configure
sudo cp /opt/telegram-daemon/config.yaml.example /opt/telegram-daemon/config.yaml
sudo nano /opt/telegram-daemon/config.yaml

# Start the service
sudo systemctl start telegram-daemon
sudo systemctl enable telegram-daemon

# Check status
sudo systemctl status telegram-daemon

# View logs
sudo journalctl -u telegram-daemon -f
```

#### Option 3: Docker

```bash
# Copy and configure
cp config.yaml.example config.yaml
nano config.yaml

# Build and run with Docker Compose
docker-compose up -d

# View logs
docker-compose logs -f
```

## First Run

On the first run, you'll be prompted to enter the verification code sent to your phone. The session will be saved to avoid re-authentication on subsequent runs.

**Note**: For automated deployment, you may need to run the daemon interactively once to complete the initial authentication, then restart it in daemon mode.

## Usage Examples

### Basic Usage

```bash
export APP_ID=your_app_id
export APP_HASH=your_app_hash
export PHONE=+1234567890
export RESPONSE_MSG="I'm no longer using Telegram. Contact me via email."

./telegram-daemon
```

### With Configuration File

```bash
./telegram-daemon config.yaml
```

### With Custom Log File

```bash
export LOG_FILE="/var/log/telegram-daemon.log"
./telegram-daemon
```

## How It Works

1. **Authentication**: Connects to Telegram using your credentials
2. **Contact Loading**: Fetches your contact list from Telegram
3. **Message Monitoring**: Listens for incoming private messages
4. **Filtering**: Ignores messages from users not in your contact list
5. **Auto-Response**: Sends predefined message to contacts (with rate limiting)
6. **Rate Limiting**: Prevents spam by limiting responses to once per configured timeout period

## Security Considerations

- Keep your `APP_ID`, `APP_HASH`, and session file secure
- Use a dedicated Telegram account if possible
- Monitor logs for unexpected activity
- Consider running in a restricted environment (container, dedicated user)

## Troubleshooting

### Common Issues

1. **Authentication Failed**: Check your APP_ID, APP_HASH, and phone number
2. **Permission Denied**: Ensure proper file permissions for session file and logs
3. **Network Issues**: Check internet connectivity and firewall settings

### Debug Mode

Enable debug logging to see detailed information:

```bash
export LOG_LEVEL=debug
./telegram-daemon
```

### Logs

- **Stdout**: Default logging destination
- **File**: Set `LOG_FILE` environment variable or `log_file` in config
- **Systemd**: Use `journalctl -u telegram-daemon -f`
- **Docker**: Use `docker-compose logs -f`

## Dependencies

- `github.com/gotd/td` - Telegram client library for Go
- `gopkg.in/yaml.v3` - YAML configuration support

## License

This project is provided as-is for educational and personal use.

