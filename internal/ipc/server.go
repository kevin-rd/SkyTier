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

type Handler func(writer message.Writer, msg *message.Message)

type router struct {
	handlers map[int]Handler
}

func (r *router) Register(kind int, handler Handler) {
	r.handlers[kind] = handler
}

func (r *router) GetRouter(kind int) (Handler, bool) {
	h, ok := r.handlers[kind]
	return h, ok
}

type Server struct {
	Router *router
}

func NewServer() *Server {
	return &Server{
		Router: &router{
			handlers: make(map[int]Handler),
		},
	}
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

	w := message.NewWriter(conn)
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

		h, ok := s.Router.GetRouter(msg.Kind)
		if !ok {
			log.Printf("[ipc] unknown message kind: %d", msg.Kind)
			continue
		}

		h(w, &msg)
	}
}

func (s *Server) Register(kind int, handler Handler) {
	s.Router.Register(kind, handler)
}
