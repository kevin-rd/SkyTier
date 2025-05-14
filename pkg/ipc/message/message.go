package message

import (
	"encoding/json"
	"errors"
	"fmt"
	"kevin-rd/my-tier/internal/peer"
	"net"
)

type Message struct {
	Version string `json:"version"`
	Kind    int    `json:"kind"`
	Payload any    `json:"payload"`
}

func (m *Message) Decode(bytes []byte) error {
	if err := json.Unmarshal(bytes, m); err != nil {
		return errors.Join(errors.New("decode error"), err)
	}

	return nil
}

const (
	KindCommand = iota
	KindPeers
)

type PeersResp struct {
	Peers []*peer.Peer `json:"peers"`
}

type Writer interface {
	Write(message *Message) error
}

type writer struct {
	net.Conn
}

func NewWriter(conn net.Conn) Writer {
	return &writer{
		Conn: conn,
	}
}

func (w *writer) Write(msg *Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return errors.Join(fmt.Errorf("marshal message error"), err)
	}
	_, err = w.Conn.Write(data)
	if err != nil {
		return err
	}
	return nil
}
