package core

import (
	"fmt"
	"io"
	"kevin-rd/my-tier/internal/router"
	"kevin-rd/my-tier/pkg/packet"
	"log"
	"net"
	"sync"
	"time"
)

var (
	bufPool = sync.Pool{
		New: func() any {
			return make([]byte, 1500)
		},
	}
)

type conn struct {
	server     *UDPServer
	remoteAddr string
	net.Conn

	buf []byte
}

func (c *conn) serve() {
	//defer func() {
	//	if err := recover(); err != nil {
	//		log.Printf("mixed_server: panic serving %v: %v", c.remoteAddr, err)
	//	}
	//	c.close()
	//}()

	// todo: 使用bufio
	c.buf = bufPool.Get().([]byte)
	for {
		n, err := c.Read(c.buf)
		if err != nil {
			if err == io.EOF {
				log.Printf("[udp_server] connection closed by client: %v", c.remoteAddr)
			} else {
				log.Printf("[udp_server] udp server read error from %v: %v", c.remoteAddr, err)
			}
			return
		}

		pkt := &packet.Packet[packet.Packable]{}
		if err = pkt.Decode(c.buf[:n]); err != nil {
			log.Printf("[udp_server] udp server read error from %v: %v", c.remoteAddr, err)
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

type UDPServer struct {
	ListenAddr string
	router     *router.Router
}

func (s *UDPServer) ListenAndServe() error {
	listener, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return fmt.Errorf("[udp_server] listen error: %w", err)
	}

	var tempDelay time.Duration // how long to sleep on accept temporary failure

	for {
		var rwc net.Conn
		rwc, err = listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
					if tempDelay > 1*time.Second {
						tempDelay = 1 * time.Second
					}
				}

				log.Println("[udp_server] temporary listen error:", err)
				time.Sleep(tempDelay)
				continue
			}
			log.Println("[udp_server] accept error:", err)
			return err
		}

		tempDelay = 0
		c := s.newConn(rwc)
		// todo: add context
		go c.serve()
	}
}

func (s *UDPServer) newConn(rwc net.Conn) *conn {
	return &conn{
		server:     s,
		remoteAddr: rwc.RemoteAddr().String(),
		Conn:       rwc,
	}
}
