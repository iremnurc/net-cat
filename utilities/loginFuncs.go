package utilities

import (
	"bufio"
	"net"
	"strings"
)

func ColorLoginFunc(conn net.Conn, reader *bufio.Reader, name string) (string, string) {
	for {

		for connection := range Clients {
			if connection != conn {
				if Clients[connection].name == name {
					NameLoginFunc(conn, reader)
					continue
				}
			}
		}

		conn.Write([]byte(ColorMenu))
		colorChoice, err := reader.ReadString('\n')
		if err != nil {
			conn.Close()
			return "", ""
		}
		colorChoice = strings.TrimSpace(colorChoice)
		userColor, userColorCode := GetColorByChoice(colorChoice)
		if userColor == "" {
			conn.Write([]byte("Invalid color choice, try again.\n"))
			continue
		}

		// Check color availability with a quick lock
		mu.Lock()
		colorExists := false
		for _, info := range Clients {
			if info.color == userColor {
				colorExists = true
				break
			}
		}
		mu.Unlock()

		if colorExists {
			conn.Write([]byte("Color already in use, choose a different color.\n"))
			continue
		}

		return userColor, userColorCode
	}
}

func NameLoginFunc(conn net.Conn, reader *bufio.Reader) string {
	for {
		conn.Write([]byte("[ENTER YOUR NAME]: "))
		nameInput, err := reader.ReadString('\n')
		if err != nil {
			chatLogger.Log("error", "Error reading user name: "+err.Error())
			conn.Close()
			return ""
		}

		name := strings.TrimSpace(nameInput)
		if name == "" || len(name) > 20 {
			conn.Write([]byte("Name cannot be empty or more than 20 characters.\n"))
			continue
		}

		// Check name availability with a quick lock
		mu.Lock()
		nameExists := false
		for _, info := range Clients {
			if info.name == name {
				nameExists = true
				break
			}
		}
		mu.Unlock()

		if nameExists {
			conn.Write([]byte("Username already exists, choose a different name.\n"))
			continue
		}

		return name
	}
}
