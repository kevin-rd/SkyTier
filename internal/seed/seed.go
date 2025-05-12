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

func (s *Service) GetPeers(network string) ([]peer.Peer, error) {
	peers, ok := s.peersMap[network]
	if !ok {
		return nil, errors.New("network not found")
	}
	return peers, nil
}
