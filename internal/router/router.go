package router

import (
	"kevin-rd/my-tier/internal/peer"
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
	log.Printf("[router] input packet: %v", pkt.Type)
	switch pkt.Type {
	case packet.TypeData:
		// 1. 是否转发
		// 2. 转发到其他节点
		// 3. 转发到本地
		r.toTun(pkt)
	default:
		r.manager.HandlePacket(w, pkt)
	}
}

func (r *Router) toTun(pkt *packet.Packet[packet.Packable]) {
	if err := r.tun.WritePacket(pkt); err != nil {
		log.Println("[router] TUN write error:", err)
	}
}
