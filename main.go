package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"main/connect"
	ra "main/ricart-agrawala"
	"net"
	"os"

	"google.golang.org/grpc"
)

var (
	name = flag.String("name", "<no name>", "Name of this instance")
	port = flag.String("port", "50050", "Port to listen on")
	l    = log.Default()
)

func (server *Server) CriticalSection(chEnter, chExit chan struct{}) {
	for {
		<-chEnter
		ra.Enter(server, 1 + len(server.peers))

		// Do critical stuff
		l.Println("In critical section")

		<-chExit
		ra.Exit(server)
	}
}

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
		log.Fatalf("fail to listen on port %s: %v", *port, err)
	}

	server := NewServer()
	ra.Init(server)

	// Serve grpc server in another thread as not to block user input
	go func() {
		grpcServer := grpc.NewServer()
		connect.RegisterConnectServer(grpcServer, server)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("stopped serving: %v", err)
		}
	}()

	// Put critical section in another thread as not to block user input
	chEnter := make(chan struct{})
	chExit := make(chan struct{})
	go server.CriticalSection(chEnter, chExit)

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Type 'lock' to request the lock, and 'unlock' to release it.")
	for scanner.Scan() {
		input := scanner.Text()

		// Dispatch commands
		switch input {
		case "lock":
			switch server.GetState() {
			case ra.Held:
				fmt.Println("Lock already held")
			case ra.Wanted:
				fmt.Println("Lock already wanted")
			case ra.Released:
				chEnter <- struct{}{}
			}
		case "unlock":
			switch server.GetState() {
			case ra.Wanted:
			case ra.Released:
				fmt.Println("Lock is not held")
			case ra.Held:
				chExit <- struct{}{}
			}
		default:
			fmt.Printf("unknown command '%s'\n", input)
		}
	}

	if scanner.Err() != nil {
		log.Fatalf("fail to read stdin: %v", scanner.Err())
	}
}
