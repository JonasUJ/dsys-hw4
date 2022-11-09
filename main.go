package main

import (
	"bufio"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"main/connect"
	"net"
	"os"
)

var (
	name = flag.String("name", "<no name>", "Name of this instance")
	port = flag.String("port", "50050", "Port to listen on")
	l    = log.Default()
)

func main() {
	flag.Parse()

	// Setup logging
	prefix := fmt.Sprintf("%-8s: ", *name)
	logfile := fmt.Sprintf("%s.log", *name)
	file, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		l.Fatalf("could not open file %s: %v", logfile, err)
	}
	l = log.New(file, prefix, log.Ltime)

	// We need a listener for grpc
	listener, err := net.Listen("tcp", net.JoinHostPort("localhost", *port))
	if err != nil {
		log.Fatalf("fail to listen on port %d: %v", *port, err)
	}
	defer listener.Close()

	server := &Server{peers: make(map[string]*Peer)}

	// Serve grpc server in another thread as not to block user input
	go func() {
		grpcServer := grpc.NewServer()
		connect.RegisterConnectServer(grpcServer, server)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("stopped serving: %v", err)
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Type 'lock' to request the lock, and 'unlock' to release it.")
	for scanner.Scan() {
		input := scanner.Text()

		// Dispatch commands
		switch input {
		case "lock":
		case "unlock":
		default:
			fmt.Printf("unknown command '%s'\n", input)
		}
	}

	if scanner.Err() != nil {
		log.Fatalf("fail to read stdin: %v", scanner.Err())
	}
}
