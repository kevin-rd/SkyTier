package core

import (
	"io"
	"kevin-rd/my-tier/internal/peer"
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

type TCPServer struct {
	ListenAddr  *net.TCPAddr
	router      *router.Router
	peerManager *peer.Manager
}

func (s *TCPServer) ListenAndServe() error {
	listener, err := net.ListenTCP("tcp", s.ListenAddr)
	if err != nil {
		return err
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
			log.Println("[tcp_server] accept error:", err)
			return err
		}

		tempDelay = 0
		c := s.newConn(rwc)
		// todo: add context
		go c.serve()
	}
}

func (s *TCPServer) newConn(rwc net.Conn) *conn {
	return &conn{
		Conn:        rwc,
		remoteAddr:  rwc.RemoteAddr(),
		router:      s.router,
		peerManager: s.peerManager,
	}
}

type conn struct {
	net.Conn
	remoteAddr net.Addr

	router      *router.Router
	peerManager *peer.Manager

	buf []byte
}

func (c *conn) RemoteAddr() net.Addr {
	return c.remoteAddr
}

func (c *conn) serve() {
	//defer func() {
	//	if err := recover(); err != nil {
	//		log.Printf("mixed_server: panic serving %v: %v", c.remoteAddr, err)
	//	}
	//	c.close()
	//}()

	// on Connected
	cw := packet.NewWriter(c, c.remoteAddr)
	log.Printf("[server] client connected: %v", c.remoteAddr)
	// c.peerManager.newConn(c.Conn)

	// todo: 使用bufio
	c.buf = bufPool.Get().([]byte)
	for {
		n, err := c.Read(c.buf)
		if err != nil {
			if err == io.EOF {
				log.Printf("[tcp_server] connection closed by client: %v", c.remoteAddr)
			} else {
				log.Printf("[tcp_server] read error from %v: %v", c.remoteAddr, err)
			}
			return
		}

		pkt := &packet.Packet[packet.Packable]{}
		if err = pkt.Decode(c.buf[:n]); err != nil {
			log.Printf("[tcp_server] packet decode error from %v: %v", c.remoteAddr, err)
			continue
		}

		c.router.Input(cw, pkt)
	}
}

func (c *conn) close() {
	if c.buf != nil {
		bufPool.Put(c.buf)
	}

	if err := c.Conn.Close(); err != nil {
		log.Println("close conn error on defer:", err)
	}
}
