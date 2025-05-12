package router

import (
	"kevin-rd/my-tier/internal/peer"
	"kevin-rd/my-tier/internal/seed"
	"kevin-rd/my-tier/internal/tun"
	"kevin-rd/my-tier/pkg/packet"
	"log"
)

type Router struct {
	// buffer
	outputCh chan *packet.Packet[packet.Packable]

	tun     *tun.TunDevice
	manager *peer.Manager
}

func NewRouter(tun *tun.TunDevice, manager *peer.Manager) *Router {
	return &Router{
		outputCh: make(chan *packet.Packet[packet.Packable], 1024),
		tun:      tun,
		manager:  manager,
	}
}

func (r *Router) Output(pkt *packet.Packet[packet.Packable]) {

}

func (r *Router) Input(w packet.Writer, pkt *packet.Packet[packet.Packable]) {
	// todo
	log.Printf("input packet: %v", pkt.Type)
	switch pkt.Type {
	case packet.TypeData:
		// 1. 是否转发
		// 2. 转发到其他节点
		// 3. 转发到本地
		r.toTun(pkt)
	case packet.TypeAuxPeers:
		// authentication
		// todo

		networkName := "default"
		peers, err := r.manager.GetPeers(networkName)
		if err != nil {
			return
		}

		if _, err = w.WritePayload(packet.TypeAuxPeersReply, &seed.PeersReplyPayload{
			Peers: peers,
		}); err != nil {
			log.Printf("write payload error: %v", err)
			return
		}
	case packet.TypePing:
		if p := r.manager.GetPeer(w.RemoteAddr()); p != nil {
			p.HandlePing(pkt)
		}
	case packet.TypeHandshake:
		handshake, ok := pkt.Payload.(*packet.PayloadHandshake)
		if !ok {
			log.Printf("invalid handshake payload")
			return
		}
		r.manager.Handshake(w, handshake)
	}
}

func (r *Router) toTun(pkt *packet.Packet[packet.Packable]) {
	if err := r.tun.WritePacket(pkt); err != nil {
		log.Println("TUN write error:", err)
	}
}
