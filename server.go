package main

import (
	"context"
	"fmt"
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
	_time      uint64
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

// Called by grpc
func (server *Server) SendMessage(ctx context.Context, msg *connect.Message) (*connect.Void, error) {
	ra.Recv(server, msg)

	// Peers reply by sending a message with our pid
	if server.GetPid() == msg.GetPid() {
		l.Printf("got reply")
		server.replyQueue <- ra.Reply{}
	} else {
		l.Printf("recv {%v}", msg)
		ra.Receive(server, ra.NewRequest(msg))
	}

	return &connect.Void{}, nil
}

// Called by grpc
func (server *Server) JoinNetwork(ctx context.Context, peerJoin *connect.PeerJoin) (*connect.ConnectedTo, error) {
	// Already connected to peer -> ignore
	if _, ok := server.peers[peerJoin.GetPid()]; !ok {
		client := server.ConnectClient(peerJoin.GetPort())
		server.AddPeer(NewPeer(peerJoin.GetPid(), client))
		l.Printf("connected to peer %d", peerJoin.GetPid())
		fmt.Printf("New peer connected (total: %d)\n", len(server.peers))

		// Tell the rest of the network about the new peer
		for pid, peer := range server.peers {
			var err error
			if pid != peerJoin.GetPid() {
				_, err = peer.client.JoinNetwork(ctx, peerJoin)
			} else {
				// Respond to the new peer telling them about us
				_, err = peer.client.JoinNetwork(ctx, &connect.PeerJoin{
					Pid:  server.GetPid(),
					Port: *port,
				})
			}

			if err != nil {
				l.Fatalf("fail to propagate join: %v", err)
			}
		}
	}

	return &connect.ConnectedTo{Pid: server.GetPid()}, nil
}

func (server *Server) AddPeer(peer *Peer) {
	server.peers[peer.pid] = peer
}

// Unused because we don't want to implement leaving the network :)
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
	// We need to know at what time we requested the lock.
	// So don't update the time if we are currently requesting it.
	if server.GetState() != ra.Wanted {
		l.Printf("changing time %d -> %d", server.GetTime(), time)
		server.time = time
	}
	// Save the time so we can update when we eventually get the lock.
	server._time = time
}

// Impl RicartAgrawala
func (server *Server) SetState(state ra.State) {
	prev := server.state
	l.Printf("changing state %s -> %s", server.state, state)
	server.state = state
	if prev == ra.Wanted {
		server.SetTime(server._time)
	}
}

// Impl RicartAgrawala
func (server *Server) GetState() ra.State {
	return server.state
}

// Impl RicartAgrawala
func (server *Server) Multicast(req ra.Request) chan ra.Reply {
	l.Printf("multicasting %v", req)
	ra.Send(server)
	req.SetTime(server.GetTime())

	for _, p := range server.peers {
		p.client.SendMessage(context.Background(), ToMessage(req))
	}

	return server.replyQueue
}

// Impl RicartAgrawala
func (server *Server) Queue() chan ra.Request {
	return server.queue
}

// Impl RicartAgrawala
func (server *Server) Reply(req ra.Request) {
	l.Printf("replying %v", req)
	ra.Send(server)
	req.SetTime(server.GetTime())
	p := server.peers[req.GetPid()]
	p.client.SendMessage(context.Background(), ToMessage(req))
}
