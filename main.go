package main

import (
	"fmt"
	"net-cat/utilities"
)

func main() {
	//Starting the logger
	logger, err := utilities.InitLogger()
	if err != nil {
		return
	}

	listener, port, err := utilities.CreatePort()
	if err != nil {
		fmt.Println("[USAGE]: ./TCPChat $port")
		logger.Log("error", "Failed to create port: "+err.Error())
		return
	}
	defer listener.Close()

	fmt.Printf("Server started on port " + port + "...\n")
	logger.Log("", "Server started on port "+port)

	// Start the idle timeout checker
	utilities.StartIdleTimeoutChecker()

	//infinite loop to accept all clients
	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Log("error", "Error accepting connection: "+err.Error())
			continue
		}

		go utilities.HandleClient(conn)
	}
}
