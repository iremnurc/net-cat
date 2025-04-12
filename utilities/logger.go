package utilities

import (
	"fmt"
	"log"
	"os"
)

// Logger struct holds the file logger instance
type Logger struct {
	fileLogger *log.Logger
}

// Global chat logger instance
var chatLogger *Logger

// InitLogger creates and initializes the logger
func InitLogger() (*Logger, error) {
	logFile, err := os.OpenFile("chat.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("error opening log file: %v", err)
	}

	chatLogger = &Logger{
		fileLogger: log.New(logFile, "", log.Ldate|log.Ltime),
	}

	return chatLogger, nil
}

// Log writes a message to the log file with appropriate prefix
func (l *Logger) Log(logType string, message string) {
	prefix := ""
	switch logType {
	case "error":
		prefix = "ERROR: "
	case "connection":
		prefix = "CONNECTION: "
	case "chat":
		// Remove timestamp from chat messages as it's already added by the logger
		if len(message) > 21 && message[0] == '[' {
			message = message[21:]
		}
		prefix = "CHAT: "
	case "chat - DM":
		prefix = "CHAT - DM: "
	}
	l.fileLogger.Println(prefix + message)
}
