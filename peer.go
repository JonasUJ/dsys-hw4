package main

import (
	"main/connect"
)

type Stream interface {
	Send(*connect.Message) error
	Recv() (*connect.Message, error)
}

type Peer struct {
	pid    uint32
	stream Stream
}

func NewPeer(stream Stream) *Peer {
	return &Peer{
		stream: stream,
	}
}

func (peer *Peer) OnRecv(fn func(*connect.Message)) {
	msg, err := peer.stream.Recv()
	if err != nil {
		return
	}

	// First message is identification
	peer.pid = msg.GetPid()

	for {
		msg, err := peer.stream.Recv()
		if err != nil {
			return
		}

		fn(msg)
	}
}
