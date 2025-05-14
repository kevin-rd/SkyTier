package ipc

import (
	"errors"
	"io"
	"kevin-rd/my-tier/pkg/ipc/message"
	"log"
	"net"
	"sync"
)

const Buf_Size = 2048

var bufPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, Buf_Size)
	},
}

type Server struct {
	Router *Router
}

func (s *Server) Serve(ln net.Listener) error {
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				log.Printf("[ipc] temporary accept error from %v: %v", conn.RemoteAddr(), err)
				continue
			}
			log.Printf("[ipc] listener accept error from %v: %v", conn.RemoteAddr(), err)
			return err
		}
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("[ipc] panic serving with remote %v: %v", conn.RemoteAddr(), err)
		}
		_ = conn.Close()
	}()

	buf := bufPool.Get().([]byte)
	defer bufPool.Put(buf)

	for {
		n, err := conn.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Printf("[ipc] connection closed by client: %v", conn.RemoteAddr())
			} else {
				log.Printf("[ipc] unexpected read error from %v: %v", conn.RemoteAddr(), err)
			}
			return
		}

		var msg message.Message
		if err = msg.Decode(buf[:n]); err != nil {
			log.Printf("[ipc] decode message error from %v: %v", conn.RemoteAddr(), err)
			continue
		}

		s.Router.serve(conn, &msg)
	}
}
