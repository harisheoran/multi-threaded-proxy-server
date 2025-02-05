package main

import (
	"bufio"
	"fmt"
	cache_lru "harisheoran/multithreaded-proxy-web-server/internal/cache"
	"io"
	"net"
	"net/http"
	"strings"
)

var (
	body                = "The target server is unavailable. Please try again later."
	customErrorResponse = "HTTP/1.1 502 Bad Gateway\r\n" +
		"Content-Type: text/plain\r\n" +
		"Content-Length: 57\r\n" +
		"Connection: close\r\n" +
		"\r\n" +
		body
)

func (app *app) handleConnection(connection net.Conn) {
	sema <- struct{}{}
	defer func() {
		<-sema
	}()

	defer func() {
		if r := recover(); r != nil {
			app.errorLogger.Println("recovered from panic", r)
		}
	}()

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

	// check the cache, if it is in cache return from cache else send request
	key := fmt.Sprintf("%s%s:%s", request.Host, request.URL.Path, request.Method)
	if ok, proxyItem := app.LRUCache.Get(key); ok {
		// prepare response
		responseToSend := proxyItem.Status +
			proxyItem.Header +
			"\r\n" +
			string(proxyItem.Body)

		app.infoLogger.Println("sending response from the LRU cache")
		connection.Write([]byte(responseToSend))
		return
	}

	// forward the request
	response, err := app.HandleRequestManual(request)
	if err != nil {
		app.errorLogger.Println("failed to handle the request from the client", err)
		connection.Write([]byte(customErrorResponse))
		return
	}

	// prepare the response to send to the user
	status, header, body := app.prepareManualResponse(response)
	responseToSend := status +
		header +
		"\r\n" +
		string(body)

	// send the response to user
	connection.Write([]byte(responseToSend))

	// save it in the cache
	proxyItem := cache_lru.ProxyItem{
		Status: status,
		Header: header,
		Body:   body,
	}
	app.LRUCache.Put(key, &proxyItem)

	defer response.Body.Close()
}

// handle request manually
func (app *app) HandleRequestManual(request *http.Request) (*http.Response, error) {
	// establish the connection to the target server
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

	// 3. combine all to form the request
	requestToSend := fmt.Sprintf("%s %s %s\r\n", request.Method, request.URL.Path, request.Proto) +
		fmt.Sprintf("Host: %s\r\n", request.Host) +
		headersToSend +
		"\r\n" +
		string(bodyToSend)

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

func (app *app) prepareManualResponse(response *http.Response) (string, string, []byte) {
	// preparing the response to send
	// 1. body
	body, err := io.ReadAll(response.Body)
	app.handleError("failed to read response body", err)

	// 2. headers
	var headerToSend string
	for key, value := range response.Header {
		headerToSend = headerToSend + fmt.Sprintf("%s: %s\r\n", key, strings.Join(value, ","))
	}

	// 3. status
	statusToSend := fmt.Sprintf(fmt.Sprintf("%s %s\r\n", response.Proto, response.Status))

	return statusToSend, headerToSend, body
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
