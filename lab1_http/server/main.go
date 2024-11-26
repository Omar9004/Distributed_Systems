package main

import (
	"bufio"
	"bytes"
	"fmt"
	"golang.org/x/net/netutil"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Respond to client with an error code
func respondWithErrorCode(conn net.Conn, code int) {
	response := &http.Response{
		StatusCode: code,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
	}
	// Set the connection to close
	response.Header.Set("Connection", "close")
	// Respond with some HTML
	response.Header.Set("Content-Type", "text/html; charset=utf-8")
	switch code {
	case 400:
		response.Status = "400 Bad Request"
		response.Body = io.NopCloser(strings.NewReader("<h1>Bad Request</h1>"))
	case 404:
		response.Status = "404 Not Found"
		response.Body = io.NopCloser(strings.NewReader("<h1>Not Found</h1>"))
	case 501:
		response.Status = "501 Not Implemented"
		response.Body = io.NopCloser(strings.NewReader("<h1>Not Implemented</h1>"))
	}
	err := response.Write(conn)
	if err != nil {
		println(err.Error())
		return
	}
}

// Check if the file extension is supported.
// Returns true and the extension if the extension is supported; otherwise, false and an empty string.
func isFileSupported(fileName string) (bool, string) {
	supportedExtensions := []string{"html", "txt", "gif", "jpeg", "jpg", "css"}
	fileNameComponents := strings.Split(fileName, ".")

	// Ensure there's only two components to the filename (name and extension)
	if len(fileNameComponents) == 2 {
		// check type of file
		for _, extension := range supportedExtensions {
			if fileNameComponents[1] == extension {
				return true, extension
			}
		}
	}
	return false, ""
}

// Handle POST requests by attempting to create a resource
func postHandler(conn net.Conn, req *http.Request) {
	response := &http.Response{
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
	}
	// Set the connection to close
	response.Header.Set("Connection", "close")
	fileName := req.RequestURI[1:]
	fileSupported, _ := isFileSupported(fileName)
	// Ensure file extension is supported
	if fileSupported {
		// Get full file path
		ex, err := os.Executable()
		if err != nil {
			// Exit the program
			panic(err)
		}
		exePath := filepath.Dir(ex)
		path := filepath.Join(exePath, "resources", fileName)
		// Read body (file content)
		bodyContent, err := io.ReadAll(req.Body)
		// Ensure the file was read successfully
		if err == nil {
			err := os.WriteFile(path, bodyContent, 0666)
			if err == nil {
				// Send OK response
				response.Status = "201 Created"
				response.StatusCode = 201
				err = response.Write(conn)
				if err != nil {
					// Error writing response to client (Socket might be closed)
					println(err.Error())
				}
				return
			} else {
				// Error writing file
				println(err.Error())
			}
		} else {
			// Couldn't read body content
			println(err.Error())
		}
	}
	// There was an error, or the extension is not supported
	respondWithErrorCode(conn, 400)
}

// Handle GET requests by attempting to fetch a resource
func getHandler(conn net.Conn, req *http.Request) {
	fileName := req.RequestURI[1:]
	fileSupported, extension := isFileSupported(fileName)

	// Ensure file extension is supported
	if fileSupported {
		// Get full file path
		ex, err := os.Executable()
		if err != nil {
			// Exit the program
			panic(err)
		}
		exePath := filepath.Dir(ex)
		path := filepath.Join(exePath, "resources", fileName)

		file, err := os.ReadFile(path)
		// Ensure the file was found
		if err == nil {
			// Write OK response
			fileToIo := io.NopCloser(bytes.NewReader(file))
			response := &http.Response{
				Status:     "200 OK",
				StatusCode: 200,
				Proto:      "HTTP/1.1",
				Body:       fileToIo,
				ProtoMajor: 1,
				ProtoMinor: 1,
				Header:     make(http.Header),
			}
			// Set the connection to close
			response.Header.Set("Connection", "close")
			// Tailor content type to extension type
			switch extension {
			case "html":
				response.Header.Set("Content-Type", "text/html; charset=utf-8")
			case "txt":
				response.Header.Set("Content-Type", "text/plain; charset=utf-8")
			case "gif":
				response.Header.Set("Content-Type", "image/gif")
			case "jpeg":
				response.Header.Set("Content-Type", "image/jpeg")
			case "jpg":
				response.Header.Set("Content-Type", "image/jpg")
			case "css":
				response.Header.Set("Content-Type", "text/css; charset=utf-8")
			}
			// Try to send the OK response
			err := response.Write(conn)
			if err != nil {
				// Error writing response to client (Socket might be closed)
				println(err.Error())
			}
			return
		} else {
			if os.IsNotExist(err) {
				// The file doesn't exist
				respondWithErrorCode(conn, 404)
				return
			} else {
				// There was an error reading the file
				println(err.Error())
			}
		}
	}
	// There was an error, or the extension is not supported
	respondWithErrorCode(conn, 400)
}

// Handle client connection.
func connectionHandler(conn net.Conn) {
	// Ensure we close the connection when handler exits
	defer func(conn net.Conn) {
		err := conn.Close()
		// If any errors occur, just print them
		if err != nil {
			fmt.Println("Error closing connection:", err)
		}
	}(conn)
	fmt.Println("New connection from", conn.RemoteAddr())
	reader := bufio.NewReader(conn)
	// Will stop blocking read of ReadRequest in case no answer arrives, or the answer is missing carriage return
	err := conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		// Not sure what kind of error would be produced here
		return
	}

	// get request object
	req, err := http.ReadRequest(reader)
	// Ensure there was no issue reading the requests
	if err != nil {
		fmt.Println("Error reading from connection:", err)
		// Close connection
		return
	}
	// Debug print request
	fmt.Println(req)

	if req.Method == "GET" {
		getHandler(conn, req)
	} else if req.Method == "POST" {
		postHandler(conn, req)
	} else {
		respondWithErrorCode(conn, 501)
	}
}

// Create server and listen for new client connections
func main() {
	args := os.Args[1:] // Obtain port number as first arg
	portNumString := fmt.Sprintf(":%s", args[0])
	listener, err := net.Listen("tcp", portNumString)
	if err != nil {
		fmt.Println("Error listening:", err)
		return
	}
	// Ensure we close the socket when program exits
	defer func(listener net.Listener) {
		// If any errors occur, just print them
		err := listener.Close()
		if err != nil {
			fmt.Println("Error closing listener:", err)
		}
	}(listener)
	fmt.Println("Server listening on port" + portNumString)
	keepRunning := true // Set keepRunning to false to close server
	// Limit the number of concurrent goroutines spawned by the server
	var max_connections = 10
	listener = netutil.LimitListener(listener, max_connections)
	for keepRunning {
		connection, err := listener.Accept() // Wait for client connection
		// Ensure they connected properly
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		// Go keyword starts new goroutine for the client to be handled
		go connectionHandler(connection)
	}
}
