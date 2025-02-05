package main

import (
	"flag"
	"fmt"
	cache_lru "harisheoran/multithreaded-proxy-web-server/internal/cache"
	"log"
	"net"
	"os"
)

var (
	maxClients = 20
	sema       = make(chan struct{}, maxClients)
)

func main() {
	// loggers
	infoLogger := log.New(os.Stdout, "INFO ", log.Lshortfile)
	errorLogger := log.New(os.Stderr, "ERROR ", log.Lshortfile)
	app := app{
		infoLogger:  infoLogger,
		errorLogger: errorLogger,
		LRUCache: &cache_lru.CacheList{
			Capacity: 2,
			MyMap:    make(map[string]*cache_lru.Node),
		},
	}

	// flag to pass the port at runtime
	flagPort := flag.String("flag", "9000", "main port of the multi-threaded proxy web server")
	flag.Parse()
	app.infoLogger.Printf("multi-threaded proxy server started at port: %s\n", *flagPort)

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
