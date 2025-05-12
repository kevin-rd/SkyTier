package engine

import (
	"fmt"
	"io"
	"kevin-rd/my-tier/internal/router"
	"kevin-rd/my-tier/pkg/packet"
	"log"
	"net"
	"sync"
)

var (
	bufPool = sync.Pool{
		New: func() any {
			return make([]byte, 1024)
		},
	}
)

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

		c.server.router.Input(packet.NewWriter(c.Conn), pkt)
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
	router     *router.Router
}

func (s *MixedServer) ListenAndServe() error {
	listener, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return fmt.Errorf("listen error: %w", err)
	}

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
