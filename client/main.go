package main

import (
	"log"
	"net"
)

func main() {
	connection, err := net.Dial("tcp", "127.0.0.1:9000")
	if err != nil {
		log.Println("failed to create the connection", err)
	}

	defer connection.Close()

	_, err = connection.Write([]byte("hello how are you"))
	if err != nil {
		log.Println("failed to send data", err)
	}

}
