package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {

	// Parse directory flag
	var dirFlag = flag.String("directory", ".", "directory to serve files from")
	flag.Parse()

	fmt.Printf("Running, with directory: %s \n", *dirFlag)
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	// Handle multiple connections
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(conn, *dirFlag)
	}
}

func handleConnection(conn net.Conn, dirFlag string) {

	// Keep the connection open until the application closes
	defer conn.Close()

	// Read the incoming connection into the buffer
	request, _, err := readRequest(conn)
	if err != nil {
		fmt.Println("Error reading request: ", err.Error())
	}

	// Extract path from request
	method, path, headers := parseRequest(request)

	// Define response
	response := ""

	// Map based on methods
	switch method {
	case "GET":
		response = handleGet(path, conn, headers, dirFlag)
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

func handleGet(path string, conn net.Conn, requestHeaders map[string]string, dirFlag string) (response string) {

	var header string
	var body string
	var statusResponse string

	var echoPattern = regexp.MustCompile(`^/echo/([a-zA-Z0-9/-]+)$`)
	var basePattern = regexp.MustCompile(`^/$`)
	var userAgentPattern = regexp.MustCompile(`^/user-agent$`)
	var filePattern = regexp.MustCompile(`^/files/([a-zA-Z0-9_-]+)$`)

	switch {
	case basePattern.MatchString(path):
		statusResponse = "HTTP/1.1 200 OK"
	case userAgentPattern.MatchString(path):
		body = requestHeaders["User-Agent"]
		statusResponse = "HTTP/1.1 200 OK"
		header = "Content-Type: text/plain\r\nContent-Length: " + fmt.Sprintf("%d", len(body))
	case echoPattern.MatchString(path):
		body = echoPattern.FindStringSubmatch(path)[1]
		statusResponse = "HTTP/1.1 200 OK"
		header = "Content-Type: text/plain\r\nContent-Length: " + fmt.Sprintf("%d", len(body))
	case filePattern.MatchString(path):
		fileName := filePattern.FindStringSubmatch(path)[1]
		filePath := filepath.Join(dirFlag, fileName)
		file, err := os.Open(filePath)
		if err != nil {
			fmt.Println("Error opening file: ", err.Error())
			statusResponse = "HTTP/1.1 404 Not Found"
		} else {
			fileInfo, err := file.Stat()
			if err != nil {
				fmt.Println("Error getting file info: ", err.Error())
				statusResponse = "HTTP/1.1 500 Internal Server Error"
			} else {
				fileSize := fileInfo.Size()
				fileBytes := make([]byte, fileSize)
				_, err = file.Read(fileBytes)
				if err != nil {
					fmt.Println("Error reading file: ", err.Error())
					statusResponse = "HTTP/1.1 500 Internal Server Error"
				} else {
					body = string(fileBytes)
					statusResponse = "HTTP/1.1 200 OK"
					header = "Content-Type: application/octet-stream\r\nContent-Length: " + fmt.Sprintf("%d", fileSize)
				}
			}
		}
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

func parseRequest(request []byte) (method string, path string, headers map[string]string) {
	/**
	Example of a HTTP request header:
	GET /index.html HTTP/1.1

	Host: localhost:4221
	User-Agent: curl/7.64.1
	**/

	lines := strings.Split(string(request), "\r\n")
	firstLine := lines[0]
	components := strings.Split(firstLine, " ")

	// Return headers as a map
	headers = make(map[string]string)
	for _, line := range lines[1:] {
		if line == "" {
			break
		}
		header := strings.Split(line, ":")
		headers[header[0]] = strings.Trim(header[1], " ")
	}

	return components[0], components[1], headers
}
