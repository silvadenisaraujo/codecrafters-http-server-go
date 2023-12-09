package main

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
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

	// Read the incoming connection into the buffer
	request, _, err := readRequest(conn)
	if err != nil {
		fmt.Println("Error reading request: ", err.Error())
	}

	// Extract path from request
	method, path := extractMethodPath(request)

	// Define response
	response := ""

	// Swtich case fo methods
	switch method {
	case "GET":
		response = handleGet(path, conn)
	default:
		fmt.Println("Method not supported: ", method)
	}

	// Send the response
	fmt.Println("Sending response: ", response)
	_, err = conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Error sending response: ", err.Error())
	}

	// Close the connection
	conn.Close()
}

func handleGet(path string, conn net.Conn) (response string) {

	var header string
	var body string
	var statusResponse string

	var echoPattern = regexp.MustCompile(`^/echo/([a-zA-Z0-9/-]+)$`)
	var basePattern = regexp.MustCompile(`^/$`)

	fmt.Println("Path: ", path)
	fmt.Println("Path: ", path)

	switch {
	case basePattern.MatchString(path):
		statusResponse = "HTTP/1.1 200 OK"
	case echoPattern.MatchString(path):
		body = echoPattern.FindStringSubmatch(path)[1]
		statusResponse = "HTTP/1.1 200 OK"
		header = "Content-Type: text/plain\r\nContent-Length: " + fmt.Sprintf("%d", len(body))
	default:
		fmt.Println("Path not found: ", path)
		statusResponse = "HTTP/1.1 404 Not Found"
	}

	response = statusResponse + "\r\n" + header + "\r\n\r\n" + body

	return response
}

func readRequest(conn net.Conn) (request []byte, n int, err error) {
	request = make([]byte, 1024)
	n, err = conn.Read(request)
	return request, n, err
}

func extractMethodPath(request []byte) (method string, path string) {
	/**
	Example of a HTTP request header:
	GET /index.html HTTP/1.1

	Host: localhost:4221
	User-Agent: curl/7.64.1
	**/

	lines := strings.Split(string(request), "\r\n")
	firstLine := lines[0]
	components := strings.Split(firstLine, " ")
	return components[0], components[1]
}
