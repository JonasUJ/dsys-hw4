package main

import (
	"main/connect"
	ra "main/ricart-agrawala"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Dial grpc server listening on the given port
func (server *Server) ConnectClient(port string) connect.ConnectClient {
	l.Printf("dialing %s", port)
	conn, err := grpc.Dial(net.JoinHostPort("localhost", port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		l.Fatalf("fail to dial: %v", err)
	}

	client := connect.NewConnectClient(conn)

	return client
}

func ToMessage(req ra.Request) *connect.Message {
	return &connect.Message{
		Time: req.GetTime(),
		Pid:  req.GetPid(),
	}
}
