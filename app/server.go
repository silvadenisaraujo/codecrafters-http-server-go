package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	fmt.Println("Running!")
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	// Keep the connection open until the application closes
	defer conn.Close()

	fmt.Println("Connection successful")
	// Return a 200 response
	response := "HTTP/1.1 200 OK\r\n\r\n"

	// Send the response
	_, err = conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Error sending response: ", err.Error())
	}
}
