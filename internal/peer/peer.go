package peer

import (
	"errors"
	"kevin-rd/my-tier/pkg/packet"
	"kevin-rd/my-tier/pkg/packet/payload"
	"kevin-rd/my-tier/pkg/utils"
	"log"
)

const (
	STATE_INIT byte = iota
	STATE_HANDSHAKE_SENT
	STATE_HANDSHAKE_RECEIVED
	STATE_HANDSHAKED
)

type Info struct {
	ID        string       // Node ID
	VirtualIP utils.IPMask // Virtual IP
}

type Peer struct {
	Info
	State      byte
	RemoteAddr string

	packet.Writer `json:"-"`

	outputCh chan *packet.Packet[packet.Packable]
}

// polling 是一个状态机
func (p *Peer) polling(self Info) error {
	for {

	}

}

// handshake 主动握手
func (p *Peer) handshake(self Info) error {
	var idBytes [32]byte
	copy(idBytes[:], self.ID)

	handshake := &packet.Packet[packet.Packable]{
		Type: packet.TypeHandshakeInit,
		Payload: &payload.HandshakeInitPayload{
			ID:        idBytes,
			DHCP:      true,
			VirtualIP: self.VirtualIP,
		},
	}
	_, err := p.WriteP(handshake)
	if err != nil {
		return err
	}
	p.State = STATE_HANDSHAKE_SENT

	resp, err := packet.ReadPacketOnce(p.GetConn())
	if err != nil {
		return err
	}
	relay, ok := resp.Payload.(*payload.HandshakeReplyPayload)
	if !ok {
		return errors.New("invalid payload type")
	}
	// todo: if need DHCP: relay.VirtualIP
	log.Printf("[peer] handshake reply from %s: %s", resp.SrcVIP, relay.Hello)

	// send HandshakeFinalize
	finalize := payload.StringPayload("ok")
	if _, err = p.WritePayload(packet.TypeHandshakeFinalize, &finalize); err != nil {
		log.Printf("[peer] write handshake finalize error: %v", err)
		return err
	}
	return nil
}

func (p *Peer) HandlePing(pkt *packet.Packet[packet.Packable]) {
	payload := payload.StringPayload("ping")
	if _, err := p.WritePayload(packet.TypePong, &payload); err != nil {
		log.Printf("[peer] write pong error: %v", err)
		return
	}
}

func (p *Peer) handshaked(info Info) {
	p.Info = info
	p.State = STATE_HANDSHAKED
}

type PeersReplyPayload struct {
	Peers []*Peer
}

func (p *PeersReplyPayload) Encode() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PeersReplyPayload) Decode(data []byte) error {
	//TODO implement me
	panic("implement me")
}

func (p *PeersReplyPayload) Length() int {
	//TODO implement me
	panic("implement me")
}
