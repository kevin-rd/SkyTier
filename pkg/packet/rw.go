package packet

import (
	"errors"
	"io"
	"log"
	"net"
	"sync"
)

var (
	bufPool = sync.Pool{
		New: func() any {
			return make([]byte, 1500)
		},
	}
)

type Writer interface {
	Write(pkt []byte) (int, error)
	WriteP(pkt *Packet[Packable]) (int, error)
	WritePayload(typ byte, payload Packable) (int, error)

	RemoteAddr() net.Addr
	// Deprecated
	GetConn() net.Conn
}

type writer struct {
	net.Conn
	remoteAddr net.Addr
}

func NewWriter(conn net.Conn, remoteAddr net.Addr) Writer {
	return &writer{conn, remoteAddr}
}

func (w *writer) Write(pkt []byte) (int, error) {
	switch c := w.Conn.(type) {
	case *net.UDPConn:
		if c.RemoteAddr() == nil {
			return c.WriteToUDP(pkt, w.remoteAddr.(*net.UDPAddr))
		} else {
			return c.Write(pkt)
		}
	default:
		return c.Write(pkt)
	}
}

func (w *writer) WriteP(pkt *Packet[Packable]) (int, error) {
	bts, err := pkt.Encode()
	if err != nil {
		return 0, err
	}
	return w.Write(bts)
}

func (w *writer) WritePayload(typ byte, payload Packable) (int, error) {
	return w.WriteP(&Packet[Packable]{
		Version: ProtocolVersion,
		Type:    typ,
		Payload: payload,
	})
}

// RemoteAddr may be have a bug, because the UDP conn dont support RemoteAddr() method.
func (w *writer) RemoteAddr() net.Addr {
	return w.remoteAddr
}

func (w *writer) GetConn() net.Conn {
	return w.Conn
}

type Reader interface {
	ReadPacket() (*Packet[Packable], error)
}

type reader struct {
	net.Conn
}

func (r *reader) ReadUntilEOF(f func(pkt *Packet[Packable]) error) error {
	buf := bufPool.Get().([]byte)
	defer bufPool.Put(buf)

	for {
		_, err := r.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return errors.Join(ErrPacketIncomplete, err)
		}

		pkt := &Packet[Packable]{}
		if err = pkt.Decode(buf); err != nil {
			log.Printf("packet decode error: %v", err)
			continue
		}

		if err = f(pkt); err != nil {
			log.Printf("error: %v", err)
		}
	}
}

// ReadPacketOnce reads only one Packet from the given reader.
func ReadPacketOnce(r io.Reader) (*Packet[Packable], error) {
	buf := bufPool.Get().([]byte)
	defer bufPool.Put(buf)

	_, err := r.Read(buf)
	if err != nil {
		return nil, errors.Join(ErrPacketIncomplete, err)
	}

	pkt := &Packet[Packable]{}
	if err = pkt.Decode(buf); err != nil {
		return nil, errors.Join(ErrPacketDecode, err)
	}
	return pkt, nil
}
