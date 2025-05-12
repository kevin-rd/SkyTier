package seed

import (
	"errors"
	"kevin-rd/my-tier/internal/peer"
)

type Service struct {
	peersMap map[string][]peer.Peer
}

type PeersReplyPayload struct {
	Peers []*peer.Peer
}

func (s *Service) GetPeers(network string) ([]peer.Peer, error) {
	peers, ok := s.peersMap[network]
	if !ok {
		return nil, errors.New("network not found")
	}
	return peers, nil
}
