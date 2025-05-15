package message

import (
	"encoding/json"
	"errors"
	"fmt"
	"kevin-rd/my-tier/internal/peer"
	"net"
)

type Message struct {
	Version string          `json:"version"`
	Kind    int             `json:"kind"`
	Payload json.RawMessage `json:"payload"`
}

func New(kind int, payload any) (*Message, error) {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.Join(errors.New("marshal payload error"), err)
	}

	return &Message{
		Version: Version,
		Kind:    kind,
		Payload: bytes,
	}, nil
}

func (m *Message) Decode(bytes []byte) error {
	if err := json.Unmarshal(bytes, m); err != nil {
		return errors.Join(errors.New("decode error"), err)
	}

	return nil
}

func DecodePayload[T any](msg *Message) (*T, error) {
	var payload T
	return &payload, json.Unmarshal(msg.Payload, &payload)
}

const (
	Version = "v1"
)

const (
	KindCommand = iota
	KindPeers
)

type PeersReq struct {
	Network string `json:"network,omitempty"`
}

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
