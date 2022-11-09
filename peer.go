package main

import (
	"log"
	"main/connect"
	ra "main/ricart-agrawala"
	"math/rand"
	"os"
	"time"
)

type Stream interface {
	Send(*connect.Message) error
	Recv() (*connect.Message, error)
}

type Peer struct {
	time uint64
	pid uint32
	stream Stream
	server *Server
	queue chan ra.Request
	replyQueue chan ra.Reply
	state ra.State
}

func NewPeer(stream Stream, server *Server) *Peer {
	return &Peer{
		time: 0,
		pid: uint32(os.Getpid()),
		stream: stream,
		server: server,
		queue: make(chan ra.Request, 100), // Can't have more than 100 peers
		replyQueue: make(chan ra.Reply, 100), 
	}
}

// Impl Lamport
func (peer *Peer) GetPid() uint32 {
	return peer.pid
}

// Impl Lamport
func (peer *Peer) GetTime() uint64 {
	return peer.time
}

// Impl LamportMut
func (peer *Peer) SetTime(time uint64) {
	peer.time = time
}

// Impl RicartAgrawala
func (peer *Peer) SetState(state ra.State) {
	peer.state = state
}

// Impl RicartAgrawala
func (peer *Peer) GetState() ra.State {
	return peer.state
}

// Impl RicartAgrawala
func (peer *Peer) Multicast(req ra.Request) chan ra.Reply {
	for _, p := range peer.server.peers {
		if p.GetPid() != peer.GetPid() {
			p.stream.Send(ToMessage(req))
		}
	}

	return peer.replyQueue
}

// Impl RicartAgrawala
func (peer *Peer) Queue() chan ra.Request {
	return peer.queue
}

// Impl RicartAgrawala
func (peer *Peer) Reply(req ra.Request) {
	p := peer.server.peers[req.GetPid()]
	p.stream.Send(ToMessage(req))
}

func (peer *Peer) Send(packet *tcp.Packet) {
	go func() {
		// Keep track of when packet is acked
		if packet.Flag != tcp.Flag_ACK {
			expected := packet.Seq + PacketLen(packet)

			log.Printf("expecting ack %d from %s", expected, peer.name)

			peer.mu.Lock()
			peer.unacked[expected] = Retransmit{
				time.Now().UnixMilli() + time.Second.Milliseconds()*2,
				packet,
			}
			peer.mu.Unlock()

			if expected != peer.LowestUnacked() {
				log.Printf("there exists unacked packets to %s, withholding %+v for now", peer.name, packet)
				return
			}
		}

		// Remember how much we've acked
		if peer.ack < packet.Ack &&
			(packet.Flag == tcp.Flag_ACK ||
				packet.Flag == tcp.Flag_SYNACK) {
			peer.ack = packet.Ack
		}

		// Chance of packet loss
		if rand.Intn(4) != 0 {
			log.Printf("[SEND] %s -> %s (%+v)\n", *name, peer.name, packet)

			// Pretend delay on the wire
			// This also simulates reordering if we send multiple packets quickly
			time.Sleep(time.Second * time.Duration(rand.Int31n(3)))

			peer.stream.Send(packet)
		} else {
			log.Printf("[SEND (lost)] %s -> %s (%+v)\n", *name, peer.name, packet)
		}
	}()
}

func (peer *Peer) Recv(chExit, chExited chan struct{}) {
	for {
		packet, err := peer.stream.Recv()
		if err != nil {
			if peer.state != Closed && peer.state != TimeWait {
				log.Printf("Recv() lost connection to %s\n", peer.name)
			}
			chExited <- struct{}{}
			return
		}

		// Exit loop when something on channel
		select {
		case <-chExit:
			return
		default:
		}

		log.Printf("[RECV] %s -> %s (%+v)\n", peer.name, *name, packet)

		peer.mu.Lock()
		if _, ok := peer.unacked[packet.Ack]; ok {
			// Packet was acked, no longer any need to save for retransmission
			log.Printf("%d was acked by %s\n", packet.Ack, peer.name)

			delete(peer.unacked, packet.Ack)
		}
		peer.mu.Unlock()

		peer.packets <- packet
	}
}

func (peer *Peer) Run() {
	chRecvExit := make(chan struct{}, 1)
	chRecvExited := make(chan struct{}, 1)
	chRetransmitExit := make(chan struct{}, 1)
	go peer.Recv(chRecvExit, chRecvExited)
	go peer.RetransmitLoop(chRetransmitExit)

	for {
		// Encodes the TCP State Machine
		switch peer.state {
		}
	}
}
