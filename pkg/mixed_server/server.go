package mixed_server

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"kevin-rd/my-tier/pkg/packet"
	"log"
	"net"
	"sync"
)

// A Request represents a request received by the server.
type Request struct {
	Conn    net.Conn       // 原始连接
	Packet  *packet.Packet // 原始 Packet
	Payload []byte         // 提取 Payload
}

// A ResponseWriter interface is used by a handler to construct a response.
type ResponseWriter interface {
	Write(pkt []byte) (int, error)
	WriteP(pkt *packet.Packet) (int, error)

	WritePayload(typ byte, payload any) (int, error)
}

type response struct {
	conn net.Conn
}

func (w *response) Write(pkt []byte) (int, error) {
	return w.conn.Write(pkt)
}

func (w *response) WriteP(pkt *packet.Packet) (int, error) {
	bts, err := pkt.Encode()
	if err != nil {
		return 0, err
	}
	return w.conn.Write(bts)
}

func (w *response) WritePayload(typ byte, payload any) (int, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(payload); err != nil {
		return 0, err
	}

	return w.WriteP(&packet.Packet{
		Version: packet.ProtocolVersion,
		Type:    typ,
		Payload: buf.Bytes(),
	})
}

// A Handler responds to a request.
type Handler interface {
	Serve(w ResponseWriter, r *Request)
}

// The HandlerFunc type is an adapter to allow the use of ordinary functions as handlers.
type HandlerFunc func(w ResponseWriter, r *Request)

func (h HandlerFunc) HandlerFunc(w ResponseWriter, r *Request) {
	h(w, r)
}

// ServeMux is a request multiplexer.
type ServeMux struct {
	handlers map[byte]HandlerFunc
	mu       sync.RWMutex
}

func NewServeMux() *ServeMux {
	return &ServeMux{
		handlers: make(map[byte]HandlerFunc),
		mu:       sync.RWMutex{},
	}
}

// RegisterHandle register handle func
func (m *ServeMux) RegisterHandle(packetType byte, h HandlerFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.handlers[packetType] = h
}

func (m *ServeMux) Serve(w ResponseWriter, r *Request) {
	if h, ok := m.handlers[r.Packet.Type]; ok {
		h(w, r)
	} else {
		log.Printf("unknown packet type: %d\n", r.Packet.Type)
	}
}

type conn struct {
	server     *MixedServer
	remoteAddr string
	net.Conn

	buf []byte
}

func (c *conn) serve() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("mixed_server: panic serving %v: %v", c.remoteAddr, err)
		}

		c.close()
	}()

	// todo: 使用bufio
	c.buf = bufPool.Get().([]byte)
	for {
		n, err := c.Read(c.buf)
		if err != nil {
			if err == io.EOF {
				log.Println("client closed connection:", err)
			} else {
				log.Println("mixed server read error:", err)
			}
			return
		}

		var pkt *packet.Packet
		pkt, err = packet.Decode(c.buf[:n])
		if err != nil {
			log.Println("decode error:", err)
			continue
		}

		w := &response{conn: c}
		r := &Request{Conn: c, Packet: pkt, Payload: pkt.Payload}

		c.server.Mux.Serve(w, r)
	}
}

func (c *conn) close() {
	if c.buf != nil {
		bufPool.Put(c.buf)
	}

	if err := c.Close(); err != nil {
		log.Println("close conn error on defer:", err)
	}
}

type MixedServer struct {
	ListenAddr string
	Mux        *ServeMux
	Handler    Handler
}

func (s *MixedServer) ListenAndServe() error {
	listener, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return fmt.Errorf("listen error: %w", err)
	}

	s.Mux = &ServeMux{handlers: make(map[byte]HandlerFunc)}

	go func() {
		for {
			var rwc net.Conn
			rwc, err = listener.Accept()
			if err != nil {
				log.Println("accept error:", err)
				continue
			}
			c := s.newConn(rwc)
			// todo: add context
			go c.serve()
		}
	}()
	return nil
}

func (s *MixedServer) newConn(rwc net.Conn) *conn {
	return &conn{
		server:     s,
		remoteAddr: rwc.RemoteAddr().String(),
		Conn:       rwc,
	}
}
