package core

import (
	"kevin-rd/my-tier/internal/peer"
	"kevin-rd/my-tier/internal/router"
	"kevin-rd/my-tier/pkg/packet"
	"log"
	"net"
)

type UDPServer struct {
	ListenAddr  *net.UDPAddr
	router      *router.Router
	peerManager *peer.Manager
}

func (s *UDPServer) ListenAndServe() error {
	ln, errL := net.ListenUDP("udp", s.ListenAddr)
	if errL != nil {
		return errL
	}

	buf := make([]byte, 1500)
	for {

		n, addr, err := ln.ReadFromUDP(buf)
		if err != nil {
			log.Println("[udp_server] read error:", err)
			continue
		}

		log.Printf("[udp_server] read from %v: %v", addr, string(buf[:n]))

		pkt := &packet.Packet[packet.Packable]{}
		if err = pkt.Decode(buf[:n]); err != nil {
			log.Printf("[udp_server] packet decode error from %v: %v", addr, err)
			continue
		}

		s.router.Input(packet.NewWriter(ln, addr), pkt)
	}
}
