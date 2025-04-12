package utilities

import (
	"fmt"
	"net"
	"strings"
	"time"
)

// ANSI color codes for terminal output
const (
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Pink   = "\033[35m"
	Cyan   = "\033[36m"
	Purple = "\033[38;5;135m"
	Orange = "\033[38;5;208m"
	Teal   = "\033[38;5;51m"
	Lime   = "\033[38;5;118m"
	Reset  = "\033[0m"
	Bold   = "\033[1m"
)

// ASCII art and menus for the chat interface
const (
	LinuxLogo = `      ,------------------,
     /` + Yellow + `     Welcome to     ` + Reset + `\
     \` + Yellow + `      TCP-Chat!     ` + Reset + `/
      '------------------'
             \
              \         _nnnn_
               \       dGGGGMMb
                      @p~qp~~qMb
                      M(@||@) M|
                      @,----.JM|
                     JS^\__/  qKL
                    dZP        qKRb
                   dZP          qKKb
                  fZP            SMMb
                  HZM            MMMM
                  FqM            MMMM
                __| ".        |\dS"qML
                |    '.       | '\ \Zq
               _)      \.___.,|     .'
               \____   )MMMMMP|   .'
                    '-'       '--'
`

	ColorMenu = `╔════════════════════════════════════════════════════════╗
║  Choose your color:                                    ║
║  ` + Red + "1. Red" + Reset + `                                                ║
║  ` + Green + "2. Green" + Reset + `                                              ║
║  ` + Yellow + "3. Yellow" + Reset + `                                             ║
║  ` + Blue + "4. Blue" + Reset + `                                               ║
║  ` + Pink + "5. Pink" + Reset + `                                               ║
║  ` + Cyan + "6. Cyan" + Reset + `                                               ║
║  ` + Purple + "7. Purple" + Reset + `                                             ║
║  ` + Orange + "8. Orange" + Reset + `                                             ║
║  ` + Teal + "9. Teal" + Reset + `                                               ║
║  ` + Lime + "10. Lime" + Reset + `                                              ║
╚════════════════════════════════════════════════════════╝
Enter number (1-10): `
)

func PrintUsage(flag string) string {
	start := "\nUSAGE:\n"
	help := "* Help usage: -h or --help\n"
	rename := "* Change your name usage: -r or --rename <new name>\n"
	quit := "* Logout usage: -q or --quit\n\n"
	dm := "* For private message usage: -dm <reciever> <private message>\n"
	color := "* Change your color: -c or --color\n"
	users := "* List online users: -u or --users\n"

	switch flag {
	case "-h", "--help":
		return start + help
	case "-r", "--rename":
		return start + rename
	case "-c", "--color":
		return start + color
	case "-u", "--users":
		return start + users
	case "-dm":
		return start + dm
	case "-q", "--quit":
		return start + quit
	default:
		return start + help + rename + color + users + dm + quit
	}
}

func PrintWelcomeMessage(conn net.Conn) {

	welcomeMsg := fmt.Sprintf("\n\033[32m╔════════════════════════════════════════════════════════╗\n" +
		"║  You can start chatting now.                           ║\n" +
		"║  Use -h or --help to see available commands.           ║\n" +
		"╚════════════════════════════════════════════════════════╝\033[0m\n\n")

	conn.Write([]byte(welcomeMsg))
}

func PrintWarningMessage(conn net.Conn) {
	// Send warning message with a prompt for input
	warningMsg := "\n\033[33m╔════════════════════════════════════════════════════════╗\n" +
		"║  WARNING: You will be disconnected due to inactivity.  ║\n" +
		"║  Type anything and press Enter to remain connected.    ║\n" +
		"╚════════════════════════════════════════════════════════╝\033[0m\n> "
	conn.Write([]byte(warningMsg))
}

// Message formatting functions
// FormatJoinMessage creates a formatted string when a user joins
func FormatJoinMessage(name string) string {
	return fmt.Sprintf("[%s] %s joined the chat\n",
		time.Now().Format("2006-01-02 15:04:05"),
		name)
}

// FormatExitMessage creates a formatted string when a user leaves
func FormatExitMessage(name string) string {
	return fmt.Sprintf("[%s] %s left the chat\n",
		time.Now().Format("2006-01-02 15:04:05"),
		name)
}

// FormatChatMessage creates a formatted string for regular chat messages
func FormatChatMessage(name, msg string) string {
	return fmt.Sprintf("[%s][%s] %s\n",
		time.Now().Format("2006-01-02 15:04:05"),
		name,
		msg)
}

// FormatPrivateMessage creates a formatted string for private messages
func FormatPrivateMessage(sender, receiver, msg string, isSender bool) string {
	if isSender {
		return fmt.Sprintf("[%s][DM to %s]: %s\n",
			time.Now().Format("2006-01-02 15:04:05"),
			receiver,
			msg)
	} else {
		return fmt.Sprintf("[%s][DM from %s]: %s\n",
			time.Now().Format("2006-01-02 15:04:05"),
			sender,
			msg)
	}
}

// FormatSystemMessage formats a system message
func FormatSystemMessage(message string) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	return fmt.Sprintf("[%s] %s", timestamp, message)
}

// FormatErrorMessage formats an error message with red and bold styling for the "Error" word
func FormatErrorMessage(message string) string {
	if strings.HasPrefix(message, "Error") {
		// Find the position of the first space after "Error"
		spaceIndex := strings.Index(message, " ")
		if spaceIndex != -1 {
			// Format "Error" in red and bold, keep the rest of the message as is
			return Bold + Red + message[:spaceIndex] + Reset + message[spaceIndex:]
		}
	}
	return message
}

// GetColorByChoice returns the ANSI color code based on user selection
func GetColorByChoice(choice string) (string, string) {
	switch choice {
	case "1":
		return Red, choice
	case "2":
		return Green, choice
	case "3":
		return Yellow, choice
	case "4":
		return Blue, choice
	case "5":
		return Pink, choice
	case "6":
		return Cyan, choice
	case "7":
		return Purple, choice
	case "8":
		return Orange, choice
	case "9":
		return Teal, choice
	case "10":
		return Lime, choice
	default:
		return "", ""
	}
}
