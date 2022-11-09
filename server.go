package main

import (
	"main/connect"
)

type Server struct {
	connect.UnimplementedConnectServer
	peers map[uint32]*Peer
}

// Called by grpc when someone dials us, and that someone is a new Peer
func (s *Server) Connect(stream connect.Connect_ConnectServer) error {
	NewPeer(stream, s).Run()

	return nil
}

func (s *Server) AddPeer(peer *Peer) {
	s.peers[peer.GetPid()] = peer
}

func (s *Server) RemovePeer(peer *Peer) {
	delete(s.peers, peer.GetPid())
}
