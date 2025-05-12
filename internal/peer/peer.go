package peer

import (
	"kevin-rd/my-tier/internal/tun"
	"kevin-rd/my-tier/pkg/packet"
	"log"
	"net"
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
	ListenAddr string

	*packet.Writer

	outputCh chan *packet.Packet
}

// polling 是一个状态机
func (p *Peer) polling(self Info) error {
	for {
		switch p.State {
		case STATE_INIT:
			if err := p.handshake(self); err != nil {
				log.Println("handshake error:", err)
				continue
			}
			p.State = STATE_HANDSHAKED
		case STATE_HANDSHAKED:
			pkt := <-p.outputCh
			_, err := p.WriteP(pkt)
			if err != nil {
				log.Println("write packet error:", err)
				continue
			}

			// todo? 是否要将接受数据和发送数据逻辑整合到一起
		}
	}

}

func (p *Peer) handshake(self Info) error {
	var idBytes [32]byte
	copy(idBytes[:], self.ID)
	payload, err := (&PayloadHandshake{
		ID:     idBytes,
		DHCP:   true,
		IpCidr: self.CIDR,
	}).MarshalBinary()
	if err != nil {
		return err
	}

	handshake := &packet.Packet{
		Type:    packet.TypeHandshake,
		Payload: payload,
	}
	_, err = p.WriteP(handshake)
	if err != nil {
		return err
	}
	return nil
}

func handleConn(conn net.Conn, t *tun.TunDevice) {
	defer conn.Close()

	go func() {
		// Peer → TUN
		buf := make([]byte, 1500)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				log.Println("read error from peer", err)
				return
			}
			err = t.WritePacket(buf[:n])
			if err != nil {
				log.Println("TUN write error:", err)
			}
		}
	}()

	// TUN → Peer
	for {
		packet, err := t.ReadPacket()
		if err != nil {
			log.Println("tun read error:", err)
			return
		}
		if _, err := conn.Write(packet); err != nil {
			log.Println("conn write error:", err)
			return
		}
	}
}
