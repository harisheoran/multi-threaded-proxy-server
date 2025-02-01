package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

func main() {
	infoLogger := log.New(os.Stdout, "INFO ", log.Lshortfile)
	errorLogger := log.New(os.Stderr, "ERROR ", log.Lshortfile)
	app := app{
		infoLogger:  infoLogger,
		errorLogger: errorLogger,
	}

	// flag to pass the port at runtime
	flagPort := flag.String("flag", "9000", "main port of the multi-threaded proxy web server")
	flag.Parse()
	log.Printf("multi-threaded proxy server started at port: %s\n", *flagPort)

	// start a socket listener at port
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", *flagPort))
	app.handleError("failed to create listener", err)

	for {
		// accept the connection
		connection, err := listener.Accept()
		app.handleError("failed to accepet connection", err)

		go app.handleConnection(connection)
	}

}

func (app *app) handleConnection(connection net.Conn) {
	defer connection.Close()
	// read the request
	reader := bufio.NewReader(connection)
	request, err := http.ReadRequest(reader)
	if err != nil {
		app.errorLogger.Println("Failed to parse the request:", err)
		connection.Close()
		return
	}

	// check the validity of the request

	// forward the request
	response, err := app.HandleRequestManual(request)
	if err != nil {
		app.errorLogger.Println("failed tp handle the request from the client", err)
		return
	}
	app.handleError("failed to forward the request", err)

	response.Write(connection)
	defer response.Body.Close()
}

func (app *app) HandleRequest(request *http.Request) (*http.Response, error) {
	newRequest, err := http.NewRequest(request.Method, request.URL.String(), request.Body)
	if err != nil {
		return nil, err
	}

	newRequest.Header = request.Header

	client := &http.Client{}
	response, err := client.Do(newRequest)
	if err != nil {
		return nil, err
	}

	app.infoLogger.Println("Request forwarded successfully")
	return response, nil

}

// handle request manually
func (app *app) HandleRequestManual(request *http.Request) (*http.Response, error) {

	// create connection
	connection, err := net.Dial("tcp", request.Host)
	if err != nil {
		return nil, err
	}

	// prepare request to send
	// 1. headers
	var headersToSend string
	for key, value := range request.Header {
		headersToSend = headersToSend + fmt.Sprintf("%s: %s\r\n", key, strings.Join(value, ","))
	}
	// 2. body
	bodyToSend, err := io.ReadAll(request.Body)
	if err != nil {
		return nil, err
	}

	// 3. combine all request
	requestToSend := fmt.Sprintf("%s %s %s\r\n", request.Method, request.URL.Path, request.Proto) +
		fmt.Sprintf("Host: %s\r\n", request.Host) +
		headersToSend +
		"\r\n" +
		string(bodyToSend)

		//	fmt.Println("FULL REQUEST:\n", requestToSend)

	// send the request
	_, err = connection.Write([]byte(requestToSend))
	if err != nil {
		return nil, err
	}

	// read the response
	reader := bufio.NewReader(connection)
	response, err := http.ReadResponse(reader, request)
	if err != nil {
		return nil, err
	}

	defer connection.Close()

	app.infoLogger.Println("request forwared and recieved successfully")

	return response, nil
}

func (app *app) prepareManualResponse(response *http.Response) {
	// preparing the response to send
	// 1. body
	body, err := io.ReadAll(response.Body)
	app.handleError("failed to read response body", err)

	// 2. headers
	var headerToSend string
	for key, value := range response.Header {
		headerToSend = headerToSend + fmt.Sprintf("%s: %s\r\n", key, strings.Join(value, ","))
	}

	// 3. final response
	responseToSend := fmt.Sprintf("%s %s\r\n", response.Proto, response.Status) +
		headerToSend

	fmt.Println(body, responseToSend)
}
