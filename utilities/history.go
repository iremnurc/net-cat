package utilities

import (
	"net"
	"strings"
)

// Maximum number of messages to keep in history
const MaxHistorySize = 20

// Stores chat history
var messageHistory []string

// AddToHistory adds a message to the chat history, maintaining the maximum size
func AddToHistory(msg string) {
	// Clean the message
	cleanMsg := strings.TrimSpace(msg)

	// Add the message to history
	messageHistory = append(messageHistory, cleanMsg)

	// If we exceed the maximum size, remove the oldest messages
	if len(messageHistory) > MaxHistorySize {
		// Remove the oldest message (first element)
		messageHistory = messageHistory[len(messageHistory)-MaxHistorySize:]
	}
}

// SendMessageHistory sends chat history to new clients
func SendMessageHistory(conn net.Conn) {
	mu.Lock()
	defer mu.Unlock()

	// Only show chat history if there are messages
	if len(messageHistory) > 0 {
		// Add chat history header
		conn.Write([]byte("\nChat History:\n"))

		for _, msg := range messageHistory {
			// Send the message with proper formatting
			conn.Write([]byte(strings.TrimSpace(msg) + "\n"))
		}
	}
}
