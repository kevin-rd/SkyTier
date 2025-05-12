package engine

import (
	"kevin-rd/my-tier/internal/seed"
	"kevin-rd/my-tier/pkg/mixed_server"
	"kevin-rd/my-tier/pkg/packet"
	"log"
)

// HandleGetPeers handle the get peers request
func (e *Engine) HandleGetPeers(w mixed_server.ResponseWriter, r *mixed_server.Request) {
	// authentication
	// todo

	networkName := "default"
	peers, err := e.seed.GetPeers(networkName)
	if err != nil {
		return
	}

	if _, err = w.WritePayload(packet.TypeAuxPeersReply, &seed.PeersReplyPayload{
		Peers: peers,
	}); err != nil {
		log.Printf("write payload error: %v", err)
		return
	}
}

func (e *Engine) HandlePing(w mixed_server.ResponseWriter, r *mixed_server.Request) {
	if _, err := w.WritePayload(packet.TypePong, "Pong"); err != nil {
		log.Printf("write pong error: %v", err)
		return
	}
}

func (e *Engine) HandleHandshake(w mixed_server.ResponseWriter, r *mixed_server.Request) {
	pkt := r.Packet
	e.peerManager.Handshake(r.Conn, pkt)
}
