# gotd Library Research

## Overview
`gotd` is a Telegram client library for Go, implementing the MTProto API. It is designed to be fast, robust, and feature-rich, with low memory overhead and automatic re-connects.

## Key Features:
- Full MTProto 2.0 implementation
- Pluggable session storage
- Automatic re-connects with keepalive
- Vendored Telegram public keys
- Rigorously tested
- 2FA support
- MTProxy support
- Helpers for uploads, downloads, messages, and queries
- Middleware for rate limiting and FLOOD_WAIT handling
- Graceful request cancellation
- WebSocket transport support

## Usage (from GitHub README):
```go
package main

import (
	"context"

	"github.com/gotd/td/telegram"
)

func main() {
	// https://core.telegram.org/api/obtaining_api_id
	client := telegram.NewClient(appID, appHash, telegram.Options{})
	if err := client.Run(context.Background(), func(ctx context.Context) error {
		// It is only valid to use client while this function is not returned
		// and ctx is not cancelled.
		api := client.API()

		// Now you can invoke MTProto RPC requests by calling the API.
		// ...

		// Return to close client connection and free up resources.
		return nil
	}); err != nil {
		panic(err)
	}
	// Client is closed.
}
```

## Obtaining `appID` and `appHash`:
To interact with the Telegram API, `appID` and `appHash` are required. These can be obtained from [my.telegram.org](https://my.telegram.org/).

## Message Handling and User Information:
From the features list, `gotd` provides helpers for `messages` and `queries`. This suggests that it can be used to listen for messages and query user information. Further investigation into the `gotd/td/telegram` package and its API methods will be necessary to understand how to:
- Listen for incoming messages.
- Identify the sender of a message.
- Determine if a sender is a contact.
- Send messages in response.

## Next Steps:
- Explore the `gotd/td/telegram` package documentation for specific API calls related to message handling and user management.
- Look for examples demonstrating message reception and sending.
- Investigate how to access contact lists or check if a user is a contact.




## `gotd/td/tg` Package Documentation

The `tg` package contains generated types from the MTProto Telegram API. This package is crucial for interacting with the Telegram API at a low level. The documentation is extensive, and it appears to be a direct mapping of the Telegram API methods and types.

### Key areas to investigate within the `tg` package:

- **Messages:** Look for types and methods related to sending, receiving, and managing messages. This will likely involve types like `Message`, `UpdateNewMessage`, `InputPeer`, etc.
- **Users and Contacts:** Identify types and methods for retrieving user information, checking contact status, and managing contacts. This might involve types like `User`, `InputUser`, `Contacts`, `ContactLink`, etc.
- **Updates:** Understand how to handle incoming updates from the Telegram server, which will include new messages, user status changes, etc.

### Initial thoughts on implementation:

1.  **Authentication:** The `telegram.NewClient` and `client.Auth().Bot` (for bot tokens) or `auth.NewFlow` (for user accounts) methods will be used for authentication.
2.  **Listening for messages:** The `client.Run` method takes a `func(ctx context.Context) error` as an argument. This function will likely be where we set up a handler for incoming updates, including new messages.
3.  **Identifying contacts:** We'll need to explore the `tg` package for methods to retrieve the user's contact list or to check if a specific user ID is in the contact list. This might involve `contacts.GetContacts` or similar methods.
4.  **Ignoring non-contacts:** Once we can identify contacts, we can filter out messages from non-contacts.
5.  **Responding to contacts:** We'll use methods within the `tg` package (e.g., `messages.SendMessage`) to send the predefined message to contacts.

## Next Steps:
- Dive deeper into the `tg` package documentation to find specific methods for message handling, user/contact management, and update processing.
- Look for examples within the `gotd` repository that demonstrate these functionalities.





## `tg` Package Structure and Key Areas for Daemon Development

The `tg` package is a comprehensive, auto-generated Go client for the Telegram MTProto API. Its extensive nature, with numerous generated files (e.g., `tl_messages_gen.go`, `tl_users_gen.go`, `tl_contacts_gen.go`), indicates a direct mapping to the Telegram API methods and types. This structure means that direct interaction with the Telegram API will involve using types and methods defined within this package.

For our daemon, the critical areas within the `tg` package will be:

*   **`messages`**: This sub-package or set of generated files will contain types and methods for sending and receiving messages. We'll need to identify how to:
    *   Handle incoming `Update` objects, specifically those related to new messages (e.g., `UpdateNewMessage`).
    *   Extract message content, sender information, and chat details from these updates.
    *   Use methods like `messages.SendMessage` to compose and send replies.

*   **`users`**: This area will provide methods to retrieve detailed information about users, such as their `ID`, `first_name`, `last_name`, and `phone` number. This will be crucial for identifying contacts.

*   **`contacts`**: This section will be vital for managing and querying the user's contact list. We'll need to find methods to:
    *   Fetch the current list of contacts (e.g., `contacts.GetContacts`).
    *   Check if a given user ID or phone number corresponds to an existing contact.

### Implications for Daemon Logic:

1.  **Authentication and Connection**: The `telegram.NewClient` and `client.Run` methods will establish the connection and handle updates. The `client.Auth()` methods will be used for user authentication.
2.  **Update Handling**: We will need to implement a mechanism to process incoming updates from Telegram. This will likely involve a loop that listens for updates and dispatches them to appropriate handlers based on their type.
3.  **Contact Identification**: Before responding to a message, the daemon will need to check if the sender is in the user's contact list. This will involve querying the `contacts` methods.
4.  **Message Filtering**: Messages from non-contacts will be ignored based on the contact identification step.
5.  **Auto-Response**: For messages from contacts, the daemon will compose and send a predefined message using the `messages` methods.

## Next Steps:

Now that we have a better understanding of the `tg` package, the next crucial step is to find practical examples that demonstrate how to use these components. I will search for `gotd` examples on GitHub and other Go-related resources, specifically looking for examples that cover:

*   Connecting to Telegram and handling updates.
*   Sending and receiving messages.
*   Retrieving user and contact information.





## Analysis of `gotd/td/examples/updates/main.go`

This example provides a solid foundation for building our Telegram daemon. It demonstrates:

*   **Client Initialization**: How to create a `telegram.Client` using `telegram.ClientFromEnvironment` which reads `APP_ID`, `APP_HASH`, `SESSION_FILE`, and `SESSION_DIR` from environment variables. This is a good practice for configuration.
*   **Authentication Flow**: The use of `auth.NewFlow` and `client.Auth().IfNecessary` for handling user authentication, including prompting for a code and 2FA password. The `examples.Terminal{}` suggests a terminal-based input for authentication, which we'll need to adapt for a daemon running on a VPS.
*   **Update Handling**: The core of the example is the `updates.New` and `tg.UpdateDispatcher` which allows registering handlers for different types of Telegram updates. Specifically, `d.OnNewChannelMessage` and `d.OnNewMessage` are crucial for our daemon.
*   **Logging**: The example uses `go.uber.org/zap` for logging, which is a good choice for structured and performant logging in a daemon.

### Key Learnings for Daemon Development:

1.  **Authentication**: We will need to implement a non-interactive authentication method suitable for a daemon. This might involve pre-registering the session or using a bot token if the user is willing to use a bot.
2.  **Message Reception**: The `d.OnNewMessage` handler is where we will implement the logic to process incoming private messages. The `tg.Entities` argument to the handler will be important for getting sender information.
3.  **Contact Identification**: The example doesn't explicitly show how to check if a sender is a contact. We will need to integrate `contacts.GetContacts` or similar `tg` package methods within the `OnNewMessage` handler.
4.  **Message Sending**: While not directly shown in this example, the `telegram/message` package (as hinted in other search results) will likely provide the functionality to send messages.

## Refined Next Steps:

- **Authentication Strategy**: Decide on the best authentication strategy for a daemon (user account vs. bot account). If a user account is chosen, we need to figure out how to handle the initial authentication step (e.g., by running it once interactively and saving the session).
- **Contact List Retrieval**: Implement the logic to fetch the user's contact list using `gotd`.
- **Message Filtering Logic**: Develop the code to filter incoming messages based on whether the sender is a contact.
- **Auto-Response Implementation**: Write the code to send a predefined message to contacts.
- **Project Structure**: Start outlining the Go project structure for the daemon.


