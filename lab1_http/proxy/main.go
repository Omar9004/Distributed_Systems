package main

import (
	"bufio"
	"fmt"
	"golang.org/x/net/netutil"
	"io"
	"net"
	"net/http"
	"os"
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
	case 501:
		response.Status = "501 Not Implemented"
		response.Body = io.NopCloser(strings.NewReader("<h1>Not Implemented</h1>"))
	case 502:
		response.Status = "502 Bad Gateway"
		response.Body = io.NopCloser(strings.NewReader("<h1>Bad Gateway</h1>"))
	}
	err := response.Write(conn)
	if err != nil {
		println(err.Error())
		return
	}
}

// Handle GET requests by attempting to fetch a resource
func getHandler(conn net.Conn, req *http.Request) {
	// Delete the proxy details from the request
	req.Header.Del("Proxy-Connection")
	// RoundTrip method, it opens a connection to the server, sending the headers and body,
	// receiving the response, and returning it.
	response, err := http.DefaultTransport.RoundTrip(req)
	// Ensure that a response was received correctly
	if err == nil {
		// Send response to client
		err = response.Write(conn)
		// Ensure client received response
		if err != nil {
			// Something went wrong when writing to client, print it
			println(err.Error())
		}
		// Close the response body
		err = response.Body.Close()
		if err != nil {
			// Something went wrong when closing body, print it
			println(err.Error())
		}
	} else {
		// Something went wrong when forwarding the request to the server
		respondWithErrorCode(conn, 502)
	}
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
	} else {
		respondWithErrorCode(conn, 501)
	}
}

// Create proxy and listen for new client connections
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
	fmt.Println("Proxy listening on port" + portNumString)
	keepRunning := true // Set keepRunning to false to close proxy
	// Limit the number of concurrent goroutines spawned by the proxy
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
