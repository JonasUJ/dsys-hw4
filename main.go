package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"main/connect"
	ra "main/ricart-agrawala"
	"net"
	"os"
	"strings"

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
		l.Printf("attempting to enter critical section")
		ra.Enter(server, 1+len(server.peers))

		l.Printf("entered critical section at time %d", server.GetTime())

		// Do critical stuff
		fmt.Println("In critical section")

		<-chExit
		l.Printf("leaving critical section")
		ra.Exit(server)
		l.Printf("left critical section")
	}
}

func main() {
	flag.Parse()

	// Setup logging
	prefix := fmt.Sprintf("%-8s(%d): ", *name, os.Getpid())
	logfile := fmt.Sprintf("%s.log", *name)
	file, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		l.Fatalf("could not open file %s: %v", logfile, err)
	}
	l = log.New(file, prefix, log.Ltime)

	// We need a listener for grpc
	listener, err := net.Listen("tcp", net.JoinHostPort("localhost", *port))
	if err != nil {
		l.Fatalf("fail to listen on port %s: %v", *port, err)
	}

	server := NewServer()
	ra.Init(server)

	// Serve grpc server in another thread as not to block user input
	go func() {
		grpcServer := grpc.NewServer()
		connect.RegisterConnectServer(grpcServer, server)
		if err := grpcServer.Serve(listener); err != nil {
			l.Fatalf("stopped serving: %v", err)
		}
	}()

	// Put critical section in another thread as not to block user input
	chEnter := make(chan struct{})
	chExit := make(chan struct{})
	go server.CriticalSection(chEnter, chExit)

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Type 'lock' to request the lock, and 'unlock' to release it.\nType 'connect <port>' to connect to a peer network at the given port.")
	for scanner.Scan() {
		input := strings.Split(scanner.Text(), " ")

		// Dispatch commands
		switch input[0] {
		case "lock":
			switch server.GetState() {
			case ra.Held:
				fmt.Println("Lock already held")
			case ra.Wanted:
				fmt.Println("Lock already wanted")
			case ra.Released:
				l.Printf("attempting to lock")
				select {
				case <-chEnter:
				default:
					chEnter <- struct{}{}
				}
			}
		case "unlock":
			switch server.GetState() {
			case ra.Wanted:
			case ra.Released:
				fmt.Println("Lock is not held")
			case ra.Held:
				select {
				case <-chExit:
				default:
					chExit  <- struct{}{}
				}
			}
		case "connect":
			client := server.ConnectClient(input[1])
			connectedTo, err := client.JoinNetwork(context.Background(), &connect.PeerJoin{
				Pid: uint32(os.Getpid()),
				Port: *port,
			})
			if err != nil {
				l.Fatalf("fail to join: %v", err)
			}

			l.Printf("connected to peer %d on port %s", connectedTo.GetPid(), input[1])

			server.AddPeer(NewPeer(connectedTo.GetPid(), client))
		default:
			fmt.Printf("unknown command '%s'\n", input)
		}
	}

	if scanner.Err() != nil {
		l.Fatalf("fail to read stdin: %v", scanner.Err())
	}
}
