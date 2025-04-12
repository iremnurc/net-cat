package utilities

import (
	"bufio"
	"fmt"
	"math"
	"net"
	"strings"
	"time"
)

// TODO disconnect offline users, format messages and logs

func Flags(conn net.Conn, message string) (string, string) {
	// Update last active timestamp
	UpdateLastActive(conn)

	SlicedMsg := strings.Fields(message)
	flag := SlicedMsg[0]

	// Helper function to handle command validation and error messages
	validateCommand := func(expectedArgs int, minArgs int, exactMatch bool) bool {
		valid := true

		if exactMatch && len(SlicedMsg) != expectedArgs {
			valid = false
		} else if !exactMatch && len(SlicedMsg) < minArgs {
			valid = false
		}

		if !valid {
			conn.Write([]byte("\033[A\033[2K"))
			errorMsg := FormatErrorMessage("\nError - Wrong command: " + Clients[conn].color + message + Reset)
			conn.Write([]byte(errorMsg + "\n"))
			conn.Write([]byte(PrintUsage(flag)))
		}

		return valid
	}

	switch flag {
	case "-h", "--help":
		if validateCommand(1, 1, true) {
			conn.Write([]byte("\033[A\033[2K"))
			conn.Write([]byte(PrintUsage("all")))
		}
	case "-r", "--rename":
		if validateCommand(2, 2, true) {
			conn.Write([]byte("\033[A\033[2K"))
			newName := SlicedMsg[1]
			Rename(conn, newName)
		}
	case "-q", "--quit":
		if validateCommand(1, 1, true) {
			conn.Write([]byte("\033[A\033[2K"))
			Logout(conn, Clients[conn].name)
			return "", ""
		}
	case "-dm":
		if validateCommand(3, 3, false) {
			reciever := SlicedMsg[1]
			scrtMsg := strings.Join(SlicedMsg[2:], " ")
			conn.Write([]byte("\033[A\033[2K"))
			PrivateMessage(reciever, scrtMsg, conn)
		}
	case "-c", "--color":
		if validateCommand(1, 1, true) {
			conn.Write([]byte("\033[A\033[2K"))
			ChangeColor(conn)
		}
	case "-u", "--users":
		if validateCommand(1, 1, true) {
			conn.Write([]byte("\033[A\033[2K"))
			ListOnlineUsers(conn)
		}
	default:
		return Clients[conn].name, message
	}
	return "", ""
}

func Logout(conn net.Conn, name string) {

	mu.Lock()
	_, exists := Clients[conn]
	if !exists {
		mu.Unlock()
		return // Client already removed, avoid crashing
	}

	// Get the IP address before deleting
	ipAddr := conn.RemoteAddr().(*net.TCPAddr).IP.String()

	// Remove the client from tracking
	delete(Clients, conn)
	delete(remoteAddresses, ipAddr)
	mu.Unlock()

	// Broadcast the exit message with the client's color
	BroadCast(conn, FormatExitMessage(name), true)

	// Add a goodbye message to the client
	conn.Write([]byte("\nYou have left the chat. Goodbye!\n"))

	// Force close the connection with a small delay to ensure messages are sent
	go func() {
		time.Sleep(time.Millisecond)
		conn.Close()
	}()
}

// Rename changes a user's name
func Rename(conn net.Conn, newName string) {
	mu.Lock()
	defer mu.Unlock()

	// Check if the new name is valid
	if newName == "" || len(newName) > 20 {
		conn.Write([]byte("Name cannot be empty or more than 20 characters\n"))
		return
	}

	// Check if the new name already exists
	var nameExists bool

	// Store the old name for the announcement
	oldName := Clients[conn].name

	for connection, info := range Clients {
		if info.name == newName && connection != conn || newName == oldName {
			conn.Write([]byte("Username: " + newName + " already exists, choose a different name\n"))
			nameExists = true
			break
		}
	}

	if nameExists {
		return
	}

	// Update the name
	Clients[conn].name = newName

	// Notify the user
	conn.Write([]byte("You have changed your name to " + newName + "\n"))

	// Announce the name change to all users
	nameChangeMsg := FormatSystemMessage(oldName + " changed their name to " + Clients[conn].color + newName + Reset)

	chatLogger.Log("chat", "User "+oldName+" "+IpAddr+" has changed their name to "+newName)

	// Add to history
	AddToHistory(nameChangeMsg)

	// Broadcast to all clients
	for client := range Clients {
		if client != conn { // Optional: don't send to the user who changed their name
			client.Write([]byte(nameChangeMsg + "\n"))
		}
	}
}

func PrivateMessage(reciever, msg string, conn net.Conn) {
	sender := Clients[conn]
	var recieverConn net.Conn

	for client := range Clients {
		if conn != client {
			if Clients[client].name == reciever {
				recieverConn = client

				goto Found // If receiver exists, skip error message
			}
		}
	}
	conn.Write([]byte(FormatErrorMessage("\nError: User not found.") + "\n"))
	return

Found:
	receiverIpAddr := recieverConn.RemoteAddr().(*net.TCPAddr).IP.String()
	// Log format for DMs
	chatLogger.Log("chat - DM", "[From "+sender.name+" "+IpAddr+" to "+reciever+" "+receiverIpAddr+" ]: "+strings.TrimSpace(msg))

	// Format messages using the FormatPrivateMessage function
	receiverMsg := FormatPrivateMessage(sender.name, reciever, msg, false)
	senderMsg := FormatPrivateMessage(sender.name, reciever, msg, true)

	recieverConn.Write([]byte(sender.color + receiverMsg + Reset))
	conn.Write([]byte(sender.color + senderMsg + Reset))
}

// ChangeColor allows a user to change their color
func ChangeColor(conn net.Conn) {
	client, exists := Clients[conn]
	if !exists {
		conn.Write([]byte(FormatErrorMessage("\nError: Client not found.") + "\n"))
		return
	}

	// Show color menu
	conn.Write([]byte(ColorMenu))

	// Read user's color choice
	reader := bufio.NewReader(conn)
	colorChoice, err := reader.ReadString('\n')
	if err != nil {
		return
	}
	colorChoice = strings.TrimSpace(colorChoice)

	// Get the new color
	newColor, colorCode := GetColorByChoice(colorChoice)

	if newColor == "" {
		conn.Write([]byte(FormatErrorMessage("\nError: Invalid color choice. Your color remains unchanged.") + "\n"))
		return
	}

	// Check if color is already in use
	mu.Lock()
	for c, info := range Clients {
		if c != conn && info.colorCode == colorCode {
			mu.Unlock()
			conn.Write([]byte(FormatErrorMessage("\nError: This color is already in use. Please choose another color.") + "\n"))
			// Try again
			ChangeColor(conn)
			return
		}
	}

	// Update client's color
	client.color = newColor
	client.colorCode = colorCode
	mu.Unlock()

	// Notify the user
	conn.Write([]byte("Your color has been changed to " + newColor + "this new color" + Reset + "\n"))

	// Notify other users
	changeMsg := "User " + client.name + " changed their color to " + newColor + client.name + Reset + "\n"

	mu.Lock()
	for connection := range Clients {
		if connection != conn {
			connection.Write([]byte(changeMsg))
		}
	}
	mu.Unlock()
}

// ListOnlineUsers displays all currently connected users
func ListOnlineUsers(conn net.Conn) {
	mu.Lock()
	defer mu.Unlock()

	if len(Clients) == 0 {
		conn.Write([]byte("No users online\n"))
		return
	}

	userList := "\nOnline Users:\n"
	i := 1
	now := time.Now()
	for _, client := range Clients {
		joinedAgo := now.Sub(client.joinedAt)

		// Calculate minutes, rounding up to at least 1 minute
		minutes := int(math.Ceil(joinedAgo.Minutes()))
		if minutes == 0 {
			minutes = 1 // Ensure it's at least 1 minute
		}

		userList += fmt.Sprintf("%d. %s%s%s (joined %d minutes ago)\n",
			i, client.color, client.name, Reset, minutes)
		i++
	}

	conn.Write([]byte(userList + "\n"))
}
