package main

import (
	"main/connect"
	ra "main/ricart-agrawala"
	"os"
)

type Server struct {
	time       uint64
	pid        uint32
	queue      chan ra.Request
	replyQueue chan ra.Reply
	state      ra.State
	connect.UnimplementedConnectServer
	peers map[uint32]*Peer
}

func NewServer() *Server {
	return &Server{
		peers:      make(map[uint32]*Peer),
		pid:        uint32(os.Getpid()),
		queue:      make(chan ra.Request, 100), // Can't have more than 100 peers
		replyQueue: make(chan ra.Reply, 100),
	}
}

func (server *Server) RecvHandler(msg *connect.Message) {
	ra.Recv(server, msg)

	// Peers reply by sending a message with our pid
	if server.GetPid() == msg.GetPid() {
		server.replyQueue <- ra.Reply{}
	} else {
		ra.Receive(server, ra.NewRequest(msg))
	}
}

// Called by grpc when someone dials us, and that someone is a new Peer
func (server *Server) Connect(stream connect.Connect_ConnectServer) error {
	go NewPeer(stream).OnRecv(server.RecvHandler)

	return nil
}

func (server *Server) AddPeer(peer *Peer) {
	server.peers[peer.pid] = peer
}

func (server *Server) RemovePeer(peer *Peer) {
	delete(server.peers, peer.pid)
}

// Impl Lamport
func (server *Server) GetPid() uint32 {
	return server.pid
}

// Impl Lamport
func (server *Server) GetTime() uint64 {
	return server.time
}

// Impl LamportMut
func (server *Server) SetTime(time uint64) {
	server.time = time
}

// Impl RicartAgrawala
func (server *Server) SetState(state ra.State) {
	server.state = state
}

// Impl RicartAgrawala
func (server *Server) GetState() ra.State {
	return server.state
}

// Impl RicartAgrawala
func (server *Server) Multicast(req ra.Request) chan ra.Reply {
	ra.Send(server)

	for _, p := range server.peers {
		p.stream.Send(ToMessage(req))
	}

	return server.replyQueue
}

// Impl RicartAgrawala
func (server *Server) Queue() chan ra.Request {
	return server.queue
}

// Impl RicartAgrawala
func (server *Server) Reply(req ra.Request) {
	ra.Send(server)
	p := server.peers[req.GetPid()]
	p.stream.Send(ToMessage(req))
}
