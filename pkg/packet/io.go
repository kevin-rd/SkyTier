package packet

import (
	"bytes"
	"encoding/gob"
	"net"
)

type Writer interface {
	Write(pkt []byte) (int, error)
	WriteP(pkt *Packet) (int, error)
	WritePayload(typ byte, payload any) (int, error)

	RemoteAddr() string
}

type writer struct {
	net.Conn
}

func NewWriter(conn net.Conn) Writer {
	return &writer{conn}
}

func (w *writer) Write(pkt []byte) (int, error) {
	return w.Conn.Write(pkt)
}

func (w *writer) WriteP(pkt *Packet) (int, error) {
	bts, err := pkt.Encode()
	if err != nil {
		return 0, err
	}
	return w.Conn.Write(bts)
}

func (w *writer) WritePayload(typ byte, payload any) (int, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(payload); err != nil {
		return 0, err
	}

	return w.WriteP(&Packet{
		Version: ProtocolVersion,
		Type:    typ,
		Payload: buf.Bytes(),
	})
}

func (w *writer) RemoteAddr() string {
	return w.Conn.RemoteAddr().String()
}
