package utilities

import (
	"errors"
	"net"
	"os"
)

func CreatePort() (net.Listener, string, error) {
	port := "8989" // Default port

	// Check for invalid number of arguments
	if len(os.Args) > 2 {
		return nil, "", errors.New("invalid usage")
	}

	// Override default port if an argument is provided
	if len(os.Args) == 2 {
		port = os.Args[1]
	}

	// Attempt to create the listener
	listener, err := net.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		return nil, "", err // Return the error to the caller
	}

	return listener, port, nil
}
