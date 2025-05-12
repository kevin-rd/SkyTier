package peer

import (
	"bytes"
	"kevin-rd/my-tier/internal/utils"
	"kevin-rd/my-tier/pkg/packet"
	"log"
	"net"
)

type Manager struct {
	Info

	// peer_id -> Peer
	Peers map[string]*Peer
}

func NewManager(id, cidr string, addrs ...string) *Manager {
	m := &Manager{
		Info:  Info{ID: id, CIDR: cidr},
		Peers: make(map[string]*Peer),
	}

	for _, addr := range addrs {
		peer, err := m.connTo(addr)
		if err != nil {
			log.Printf("connect to %s error: %v", addr, err)
			continue
		}
		m.Peers[peer.ID] = peer
	}

	go m.manage()
	return m
}

func (m *Manager) manage() {
	// todo: 启动
}

// connTo 主动连接到Peer
func (m *Manager) connTo(addr string) (*Peer, error) {
	conn, err := net.Dial("udp", addr)
	if err != nil {
		log.Printf("dial %s error: %v", addr, err)
		return nil, err
	}

	return &Peer{
		Info:   Info{ID: "unknown_" + utils.RandomString(16), CIDR: addr},
		Writer: packet.NewWriter(conn),
	}, nil
}

// Handshake 处理握手消息, 被动连接Peer
func (m *Manager) Handshake(conn net.Conn, pkt *packet.Packet) {
	handshake, err := Decode(pkt.Payload)
	if err != nil {
		log.Println("decode handshake error:", err)
		return
	}

	id := bytes.Trim(handshake.ID[:], "\x00")
	peer, ok := m.Peers[string(id)]
	if !ok {
		log.Printf("new peer: %s %s", id, handshake.IpCidr)
		peer = &Peer{
			Info: Info{string(id), handshake.IpCidr},
		}
	}
	peer.Writer = packet.NewWriter(conn)
	peer.State = STATE_HANDSHAKED
}
