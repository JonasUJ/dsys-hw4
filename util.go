package main

import (
	"context"
	"log"
	"main/connect"
	"net"
	ra "main/ricart-agrawala"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Dial grpc server listening on the given port and create a new Peer
func Connect(port string, server *Server) {
	conn, err := grpc.Dial(net.JoinHostPort("localhost", port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}

	client := connect.NewConnectClient(conn)
	stream, err := client.Connect(context.Background())
	if err != nil {
		log.Fatalf("fail to connect: %v", err)
	}

	go NewPeer(stream, server).Run()
}

func ToMessage(req ra.Request) *connect.Message {
	return &connect.Message{
		Time: req.GetTime(),
		Pid: req.GetPid(),
	}
}
