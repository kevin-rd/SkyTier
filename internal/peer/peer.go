package peer

import (
	"kevin-rd/my-tier/internal/tun"
	"log"
	"net"
)

type Peer struct {
	ID         string
	ListenAddr string
	PeerAddr   string
}

func (p *Peer) Conn(t *tun.TunDevice) error {
	conn, err := net.Dial("tcp", p.PeerAddr)
	if err != nil {
		return err
	}
	log.Println("Connected to peer", p.PeerAddr)
	handleConn(conn, t)
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
