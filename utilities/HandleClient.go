package utilities

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

// Mutex for thread-safe operations
var mu sync.Mutex

// UserInfo holds client information
type UserInfo struct {
	name       string
	color      string
	colorCode  string
	joinedAt   time.Time
	lastActive time.Time
}

// Map to store active client connections
var (
	Clients = make(map[net.Conn]*UserInfo) //to keep the client/conn names dynamic we need a pointer

	// Map to track remote addresses
	remoteAddresses = make(map[string]bool)

	// Extract client IP
	IpAddr string
)

const (
	MaxUsers         = 10
	MaxMessageLength = 200
)

func HandleClient(conn net.Conn) {

	IpAddr = conn.RemoteAddr().(*net.TCPAddr).IP.String()
	// Lock only while modifying shared data
	mu.Lock()
	if len(Clients) >= MaxUsers {
		mu.Unlock()
		conn.Write([]byte("The server is full, please try again later.\n"))
		conn.Close()
		return
	}
	if remoteAddresses[IpAddr] {
		mu.Unlock()
		conn.Write([]byte("You are already connected to the chat.\n"))
		conn.Close()
		return
	}
	remoteAddresses[IpAddr] = true
	mu.Unlock()

	// Send welcome message (No need to hold lock)
	conn.Write([]byte(LinuxLogo))
	chatLogger.Log("connection", "New connection from "+conn.RemoteAddr().String())

	reader := bufio.NewReader(conn)

	name := NameLoginFunc(conn, reader)

	userColor, userColorCode := ColorLoginFunc(conn, reader, name)

	// Register client
	now := time.Now()
	mu.Lock()
	Clients[conn] = &UserInfo{
		name:       name,
		color:      userColor,
		colorCode:  userColorCode,
		joinedAt:   now,
		lastActive: now,
	}
	mu.Unlock()

	chatLogger.Log("connection", "User "+name+" "+IpAddr+" joined the chat")

	// Send chat history and welcome message
	SendMessageHistory(conn)
	PrintWelcomeMessage(conn)

	// Notify others about the new user
	joinMsg := FormatJoinMessage(name)
	mu.Lock()
	messageHistory = append(messageHistory, joinMsg)
	mu.Unlock()

	// Broadcast to other users
	go func() {
		mu.Lock()
		for client := range Clients {
			if client != conn { // Send to everyone except the new user
				client.Write([]byte(userColor + joinMsg + Reset))
			}
		}
		mu.Unlock()
	}()

	// Handle messages in a new goroutine
	go handleMessages(conn, reader, name)
}

// BroadCast sends a message to all connected clients
func BroadCast(conn net.Conn, msg string, exit bool) {
	mu.Lock()
	defer mu.Unlock()

	senderInfo, exists := Clients[conn]
	if !exists {
		senderInfo = &UserInfo{name: "Unknown", color: Reset} // Fallback if sender is gone
	}

	// Only log chat messages, not exit messages
	if !exit {
		chatLogger.Log("chat", strings.TrimSpace(msg))
	}

	// Add to message history
	AddToHistory(msg)

	for client := range Clients {
		client.Write([]byte(senderInfo.color + msg + Reset))
	}
}

// handleMessages processes incoming messages from a client
func handleMessages(conn net.Conn, reader *bufio.Reader, name string) {
	go func() {
		for {
			msg, err := reader.ReadString('\n')
			if err != nil {
				Logout(conn, name)
				chatLogger.Log("connection", "User "+name+" "+IpAddr+" disconnected")
				return
			}

			// Update last active timestamp
			UpdateLastActive(conn)

			// Check if this connection is in warning response mode
			warningMu.Lock()
			inWarning := inWarningResponse[conn]
			warningMu.Unlock()

			if inWarning {
				// Reset warning status
				warnedUsersMu.Lock()
				delete(warnedUsers, conn)
				warnedUsersMu.Unlock()

				// Confirm they're staying connected
				conn.Write([]byte("\033[32mYou will remain connected.\033[0m\n"))

				// Remove from warning response mode
				warningMu.Lock()
				delete(inWarningResponse, conn)
				warningMu.Unlock()

				continue
			}

			// Normal message processing
			msg = strings.TrimSpace(msg)

			// Skip empty messages
			if msg == "" {
				conn.Write([]byte("\033[1A\033[2K"))
				continue
			}

			name, message := Flags(conn, msg)
			if message != "" {
				if len(message) > MaxMessageLength {
					conn.Write([]byte("\033[A\033[2K"))
					conn.Write([]byte(FormatErrorMessage("Error: Message too long. Maximum length is "+fmt.Sprint(MaxMessageLength)+" characters.") + "\n"))
					continue
				}

				conn.Write([]byte("\033[A\033[2K"))
				BroadCast(conn, FormatChatMessage(name, message), false)
			}
		}
	}()
}
