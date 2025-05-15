package peer

import (
	"bytes"
	"kevin-rd/my-tier/pkg/packet"
	"kevin-rd/my-tier/pkg/utils"
	"log"
	"net"
	"sync"
)

type Manager struct {
	Info

	mu sync.Mutex
	// remote_addr -> Peer
	peerMap map[string]*Peer
	// network_name -> []*Peer
	peerGroup map[string][]*Peer
}

func NewManager(id, cidr string, addrs ...string) *Manager {
	m := &Manager{
		Info:      Info{ID: id, CIDR: cidr},
		peerMap:   make(map[string]*Peer),
		peerGroup: make(map[string][]*Peer),
	}

	// add self to peers
	m.addPeer("", &Peer{
		Info: Info{ID: id, CIDR: cidr},
		// todo: Writer, RemoteAddr
	})

	for _, addr := range addrs {
		peer, err := m.connTo(addr)
		if err != nil {
			log.Printf("connect to %s error: %v", addr, err)
			continue
		}
		m.peerMap[addr] = peer
	}

	return m
}

func (m *Manager) GetPeer(remoteAddr string) *Peer {
	if peer, ok := m.peerMap[remoteAddr]; ok {
		return peer
	}
	return nil
}

func (m *Manager) GetPeers(network string) []*Peer {
	// todo
	return m.peerGroup[network]
}

func (m *Manager) Manage() error {
	// todo: 启动
	return nil
}

// connTo 主动连接到Peer
func (m *Manager) connTo(addr string) (*Peer, error) {
	conn, err := net.Dial("udp", addr)
	if err != nil {
		log.Printf("[peer] dial %s error: %v", addr, err)
		return nil, err
	}

	return &Peer{
		Info:   Info{ID: "unknown_" + utils.RandomString(16), CIDR: addr},
		Writer: packet.NewWriter(conn),
	}, nil
}

// Handshake 处理握手消息, 被动连接Peer
func (m *Manager) Handshake(w packet.Writer, handshake *packet.PayloadHandshake) {
	id := bytes.Trim(handshake.ID[:], "\x00")
	peer, ok := m.peerMap[string(id)]
	if !ok {
		log.Printf("[peer] new peer: %s %s", id, handshake.IpCidr)
		peer = &Peer{
			Info: Info{string(id), handshake.IpCidr},
		}
	}
	// add to group
	m.addPeer("", peer)

	// write reply message
	pkt := packet.NewPacket(packet.TypeHandshakeReply, &packet.HandshakeReplyPayload{Hello: "ni hao"})
	if _, err := w.WriteP(pkt); err != nil {
		log.Printf("[peer] write handshake reply error: %v", err)
		return
	}
	peer.State = STATE_HANDSHAKED
}

func (m *Manager) addPeer(network string, peer *Peer) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.peerMap[peer.RemoteAddr] = peer
	m.peerGroup[network] = append(m.peerGroup[network], peer)
}
