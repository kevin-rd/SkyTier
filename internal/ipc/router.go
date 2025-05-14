package ipc

import (
	"kevin-rd/my-tier/pkg/ipc/message"
	"log"
	"net"
)

type Handler func(writer message.Writer, msg *message.Message)

type Router struct {
	handlers map[int]Handler
}

func NewRouter() *Router {
	return &Router{
		handlers: make(map[int]Handler),
	}
}

func (r *Router) Register(kind int, handler Handler) {
	r.handlers[kind] = handler
}

func (r *Router) serve(conn net.Conn, msg *message.Message) {
	h, ok := r.handlers[msg.Kind]
	if !ok {
		log.Printf("[ipc] unknown message kind: %d", msg.Kind)
		return
	}
	w := message.NewWriter(conn)
	h(w, msg)
}
