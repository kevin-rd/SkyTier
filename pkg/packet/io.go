package packet

import (
	"bytes"
	"encoding/gob"
	"net"
)

type Writer struct {
	conn net.Conn
}

func NewWriter(conn net.Conn) *Writer {
	return &Writer{conn: conn}
}

func (w *Writer) Write(pkt []byte) (int, error) {
	return w.conn.Write(pkt)
}

func (w *Writer) WriteP(pkt *Packet) (int, error) {
	bts, err := pkt.Encode()
	if err != nil {
		return 0, err
	}
	return w.conn.Write(bts)
}

func (w *Writer) WritePayload(typ byte, payload any) (int, error) {
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
