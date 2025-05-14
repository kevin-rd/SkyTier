package peer

import (
	"kevin-rd/my-tier/pkg/packet"
	"log"
)

const (
	STATE_INIT       = 0
	STATE_HANDSHAKED = 1
)

type Info struct {
	ID   string
	CIDR string
}

type Peer struct {
	Info
	State      int32
	RemoteAddr string

	packet.Writer

	outputCh chan *packet.Packet[packet.Packable]
}

// polling 是一个状态机
func (p *Peer) polling(self Info) error {
	for {
		switch p.State {
		case STATE_INIT:
			if err := p.handshake(self); err != nil {
				log.Println("[peer] handshake error:", err)
				continue
			}
			p.State = STATE_HANDSHAKED
		case STATE_HANDSHAKED:
			pkt := <-p.outputCh
			_, err := p.WriteP(pkt)
			if err != nil {
				log.Println("[peer] write packet error:", err)
				continue
			}

			// todo? 是否要将接受数据和发送数据逻辑整合到一起
		}
	}

}

func (p *Peer) handshake(self Info) error {
	var idBytes [32]byte
	copy(idBytes[:], self.ID)

	handshake := &packet.Packet[packet.Packable]{
		Type: packet.TypeHandshake,
		Payload: &packet.PayloadHandshake{
			ID:     idBytes,
			DHCP:   true,
			IpCidr: self.CIDR,
		},
	}
	_, err := p.WriteP(handshake)
	if err != nil {
		return err
	}
	return nil
}

func (p *Peer) HandlePing(pkt *packet.Packet[packet.Packable]) {
	payload := packet.StringPayload("ping")
	if _, err := p.WritePayload(packet.TypePong, &payload); err != nil {
		log.Printf("[peer] write pong error: %v", err)
		return
	}
}
