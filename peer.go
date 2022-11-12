package main

import (
	"main/connect"
)

type Peer struct {
	pid    uint32
	client connect.ConnectClient
}

func NewPeer(pid uint32, client connect.ConnectClient) *Peer {
	return &Peer{
		pid:    pid,
		client: client,
	}
}
