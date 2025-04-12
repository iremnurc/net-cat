package utilities

import (
	"net"
	"sync"
	"time"
)

const (
	//Just for audits
	// IdleTimeout   = 10 * time.Second
	// WarningTime   = 5 * time.Second
	// CheckInterval = 1 * time.Second

	IdleTimeout   = 10 * time.Minute
	WarningTime   = 8 * time.Minute
	CheckInterval = 2 * time.Minute
)

// Add these variables at the package level
var (
	// Mutex for protecting the inWarningResponse map
	warningMu sync.Mutex

	// Map to track which connections are currently in warning response mode
	inWarningResponse = make(map[net.Conn]bool)

	// Mutex for protecting the warnedUsers map
	warnedUsersMu sync.Mutex

	// Map to track which users have been warned
	warnedUsers = make(map[net.Conn]bool)
)

// StartIdleTimeoutChecker starts a goroutine that periodically checks for idle users
func StartIdleTimeoutChecker() {

	go func() {
		for {
			time.Sleep(CheckInterval)

			mu.Lock()
			now := time.Now()
			var idleConns []net.Conn
			var warningConns []net.Conn

			// Find idle and warning connections
			for conn, info := range Clients {
				idleTime := now.Sub(info.lastActive)

				if idleTime > IdleTimeout {
					idleConns = append(idleConns, conn)

					warnedUsersMu.Lock()
					delete(warnedUsers, conn)
					warnedUsersMu.Unlock()

				} else if idleTime > WarningTime {
					warnedUsersMu.Lock()
					alreadyWarned := warnedUsers[conn]
					warnedUsersMu.Unlock()

					if !alreadyWarned {
						warningConns = append(warningConns, conn)
					}
				} else {
					// User is active, reset their warning status
					warnedUsersMu.Lock()
					if warnedUsers[conn] {
						delete(warnedUsers, conn)
					}
					warnedUsersMu.Unlock()
				}
			}
			mu.Unlock()

			// Send warnings
			for _, conn := range warningConns {
				mu.Lock()
				_, exists := Clients[conn]
				mu.Unlock()

				if exists {
					warnedUsersMu.Lock()
					warnedUsers[conn] = true
					warnedUsersMu.Unlock()

					// Mark this connection as in warning response mode
					warningMu.Lock()
					inWarningResponse[conn] = true
					warningMu.Unlock()

					PrintWarningMessage(conn)
				}
			}

			// Disconnect idle users
			for _, conn := range idleConns {
				mu.Lock()
				info, exists := Clients[conn]
				mu.Unlock()

				if exists {

					// Notify the user
					conn.Write([]byte("\033[1;31mYou have been disconnected due to inactivity.\033[0m\n"))

					// Clean up warning response mode
					warningMu.Lock()
					delete(inWarningResponse, conn)
					warningMu.Unlock()

					// Get the IP address before deleting
					ipAddr := conn.RemoteAddr().(*net.TCPAddr).IP.String()

					mu.Lock()
					// Remove from clients map
					delete(Clients, conn)
					// Remove from remoteAddresses map
					delete(remoteAddresses, ipAddr)
					mu.Unlock()

					// Broadcast the exit message
					BroadCast(conn, FormatExitMessage(info.name), true)

					// Close the connection
					conn.Close()

					warnedUsersMu.Lock()
					delete(warnedUsers, conn)
					warnedUsersMu.Unlock()
				}
			}
		}
	}()
}

// UpdateLastActive updates the timestamp of the user's last activity
func UpdateLastActive(conn net.Conn) {
	mu.Lock()
	defer mu.Unlock()

	if client, exists := Clients[conn]; exists {
		client.lastActive = time.Now()
	}
}
