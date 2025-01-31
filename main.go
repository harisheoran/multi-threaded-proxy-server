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
		defer connection.Close()
		app.handleError("failed to accepet connection", err)

		// read the request
		reader := bufio.NewReader(connection)
		request, err := http.ReadRequest(reader)
		app.handleError("failed to prase the http request", err)

		// check the request
		isValid := app.ValidateRequest(request)
		if !isValid {
			app.infoLogger.Println("Not a valid request", request.RequestURI)
		}
		// forward the request
		response, err := app.HandleRequest(request)
		app.handleError("failed to forward the request", err)

		fmt.Println("reading the responsebody", response)

		// preparing the response to send
		// 1. body

		body, err := io.ReadAll(response.Body)
		app.handleError("failed to read response body", err)

		// 2. headers
		var headerToSend string
		for key, value := range response.Header {
			headerToSend = headerToSend + fmt.Sprintf("%s: %s\r\n", key, value[0])
		}

		// 3. final response
		responseToSend := fmt.Sprintf("%s %s\r\n", response.Proto, response.Status) +
			headerToSend

		connection.Write([]byte(responseToSend))
		connection.Write([]byte("\r\n")) // End of headers
		connection.Write(body)
		connection.Close()
	}

}

func (app *app) ValidateRequest(request *http.Request) bool {
	if request.Host == "" {
		return false
	}

	return true
}

func (app *app) HandleRequest(request *http.Request) (*http.Response, error) {
	// create connection
	connection, err := net.Dial("tcp", request.Host)
	if err != nil {
		return nil, err
	}

	// prepare request to send
	requestToSend := fmt.Sprintf("GET %s HTTP/1.1\r\n", request.URL.Path) +
		fmt.Sprintf("Host: %s\r\n", request.Host) +
		"\r\n"

	// send the request
	connection.Write([]byte(requestToSend))

	reader := bufio.NewReader(connection)
	response, err := http.ReadResponse(reader, request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	defer connection.Close()

	app.infoLogger.Println("request forwared and recieved successfully")

	return response, nil
}
