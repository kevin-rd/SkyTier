package peer

import (
	"kevin-rd/my-tier/pkg/packet"
	"kevin-rd/my-tier/pkg/utils"
	"log"
	"net"
	"sync"
	"time"
)

type Manager struct {
	Info

	mu sync.Mutex
	// unHandshake, remoteAddr -> Peer
	tempPeers map[string]*Peer
	// virtualIP -> Peer
	peerMap map[utils.IPv4]*Peer
	// network_name -> []*Peer
	peerGroup map[string][]*Peer
}

func NewManager(id string, cidr [5]byte, addrs ...string) *Manager {
	m := &Manager{
		Info:      Info{ID: id, VirtualIP: cidr},
		peerMap:   map[utils.IPv4]*Peer{},
		peerGroup: map[string][]*Peer{},
		tempPeers: map[string]*Peer{},
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// add self to peers
	m.addPeer("", &Peer{
		Info: Info{ID: id, VirtualIP: cidr},
		// todo: Writer, RemoteAddr
	})

	// conn to default peers
	for _, addr := range addrs {
		conn, err := net.Dial("udp", addr)
		if err != nil {
			log.Printf("[peer_manager] connect to %s error: %v", addr, err)
			continue
		}
		log.Printf("[peer] connect to %s success", addr)
		m.newConn(packet.NewWriter(conn, conn.RemoteAddr()))
	}

	return m
}

func (m *Manager) GetPeer(vip utils.IPv4) *Peer {
	if peer, ok := m.peerMap[vip]; ok {
		return peer
	}
	return nil
}

func (m *Manager) GetPeers(network string) []*Peer {
	// todo
	return m.peerGroup[network]
}

func (m *Manager) Manage() error {
	// process tempPeers
	for _, p := range m.tempPeers {
		switch p.State {
		case STATE_INIT:
			if err := p.handshake(m.Info); err != nil {
				log.Println("[peer] handshake error:", err)
				time.Sleep(time.Second * 1)
				continue
			}
			m.handshaked(p, p.Info)
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

	return nil
}

// newConn add new Conn to tempPeers
func (m *Manager) newConn(writer packet.Writer) *Peer {
	p, ok := m.tempPeers[writer.RemoteAddr().String()]
	if !ok {
		log.Printf("[peer] new peer: %s", writer.RemoteAddr())
		p = &Peer{
			State:      STATE_INIT,
			RemoteAddr: writer.RemoteAddr().String(),
			Writer:     writer,
			outputCh:   make(chan *packet.Packet[packet.Packable]),
		}
		m.tempPeers[writer.RemoteAddr().String()] = p
	}
	return p
}

func (m *Manager) handshaked(p *Peer, info Info) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.tempPeers, p.RemoteAddr)
	m.addPeer("", p)
	p.handshaked(info)
}

func (m *Manager) addPeer(network string, peer *Peer) {
	m.peerMap[utils.IPv4(peer.VirtualIP[:4])] = peer
	m.peerGroup[network] = append(m.peerGroup[network], peer)
}
