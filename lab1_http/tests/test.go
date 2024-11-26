package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

func otherRequest(fileName string, method string) *http.Request {
	if len(fileName) > 0 && fileName[0] != '/' {
		fileName = "/" + fileName
	}
	req := &http.Request{
		Method: method,
		URL: &url.URL{
			Scheme: "http",
			Host:   "localhost:8080",
			Path:   fileName,
		},
		Header:     make(http.Header),
		Body:       nil,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Host:       "localhost:8080",
	}
	return req
}
func getRequest(fileName string) *http.Request {
	if len(fileName) > 0 && fileName[0] != '/' {
		fileName = "/" + fileName
	}
	fmt.Println("File_name: ", fileName)
	req := &http.Request{
		Method: "GET",
		URL: &url.URL{
			Scheme: "http",
			Host:   "localhost:8080",
			Path:   fileName,
		},
		Header:     make(http.Header),
		Body:       nil,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Host:       "localhost:8080",
	}
	return req
}

func postRequest(fileName string) (*http.Request, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, file)
	if err != nil {
		return nil, err
	}
	baseFileName := filepath.Base(fileName)
	fmt.Println(baseFileName)
	req := &http.Request{
		Method: "POST",
		URL: &url.URL{
			Scheme: "http",
			Host:   "localhost:8080",
			Path:   "/" + baseFileName,
		},
		Header:     make(http.Header),
		Body:       io.NopCloser(buf),
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Host:       "localhost:8080",
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Content-Length", fmt.Sprintf("%d", buf.Len()))

	// Print the request details for debugging
	fmt.Printf("Sending POST request to: %s\n", req.URL.String())
	fmt.Printf("Content-Length: %s\n", req.Header.Get("Content-Length"))

	return req, nil
}

func main() {
	args := os.Args[1:]
	portNumString := fmt.Sprintf(":%s", args[0])
	req_type := fmt.Sprintf("%s", args[1])
	file_name := fmt.Sprintf("%s", args[2])
	
	if req_type == "POST"{
		_, err := os.Stat(file_name)
		if err != nil {
			fmt.Println("File does not exist!", err)
		}
	}
		


	//Connect to the TCP server on port the pre-determined port (e.g. 8080)
	conn, err := net.Dial("tcp", "localhost"+portNumString)

	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}

	defer conn.Close()

	fmt.Println("Connected to server")
	fmt.Println("Enter a message to send to the server:")
	message := ""

	var request *http.Request

	switch req_type {
	case "GET":
		request = getRequest(file_name)
	case "POST":
		request, err = postRequest(file_name)
		if err != nil {
			return
		}
	default:
		request = otherRequest(file_name, req_type)
	}
	request.Write(conn)

	//Read the echo/response message from the server
	buf := make([]byte, len(message))
	reader := bufio.NewReader(conn)
	res, err := http.ReadResponse(reader, request)
	fmt.Println(res)

	fmt.Println("Server response:", string(buf))

}
