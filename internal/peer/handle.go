package peer

import (
	"bytes"
	"kevin-rd/my-tier/pkg/packet"
	"kevin-rd/my-tier/pkg/packet/payload"
	"log"
)

func (m *Manager) HandlePacket(w packet.Writer, pkt *packet.Packet[packet.Packable]) {
	switch pkt.Type {
	case packet.TypeAuxPeers:
		// authentication
		// todo

		networkName := "default"
		peers := m.GetPeers(networkName)

		if _, err := w.WritePayload(packet.TypeAuxPeersReply, &PeersReplyPayload{
			Peers: peers,
		}); err != nil {
			log.Printf("[router] write payload error: %v", err)
			return
		}
	case packet.TypePing:
		if p := m.GetPeer(pkt.SrcVIP); p != nil {
			p.HandlePing(pkt)
		}
	case packet.TypeHandshakeInit:
		handshake, ok := pkt.Payload.(*payload.HandshakeInitPayload)
		if !ok {
			log.Printf("[router] invalid handshake payload")
			return
		}
		m.HandshakeInit(w, pkt, handshake)
	case packet.TypeHandshakeReply:
		panic("unhandled HandshakeReply cae")
	case packet.TypeHandshakeFinalize:
		p := m.GetPeer(pkt.SrcVIP)
		if p != nil {
			p.State = STATE_HANDSHAKED
		}
	default:
		panic("unhandled default case")
	}
}

// HandshakeInit 处理握手消息, 被动连接Peer
func (m *Manager) HandshakeInit(w packet.Writer, pkt *packet.Packet[packet.Packable], handshake *payload.HandshakeInitPayload) {
	id := bytes.Trim(handshake.ID[:], "\x00")

	peer, ok := m.tempPeers[w.RemoteAddr().String()]
	if !ok {
		log.Printf("[peer] new peer: %s %s", id, handshake.VirtualIP)
		peer = m.newConn(w)
	}

	// write reply message
	resp := packet.NewPacket(packet.TypeHandshakeReply, &payload.HandshakeReplyPayload{Hello: "ni hao"})
	if _, err := w.WriteP(resp); err != nil {
		log.Printf("[peer] write handshake reply error: %v", err)
		return
	}
	peer.State = STATE_HANDSHAKE_RECEIVED

	// send handshake finalize
	relay := payload.HandshakeReplyPayload{Hello: "ok"}
	if _, err := w.WritePayload(packet.TypeHandshakeFinalize, &relay); err != nil {
		log.Printf("[peer] write handshake finalize error: %v", err)
	}
	// handshake success
	m.handshaked(peer, Info{ID: string(id), VirtualIP: handshake.VirtualIP})
}
