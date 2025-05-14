package unix_socket

import (
	"errors"
	"kevin-rd/my-tier/internal/ipc"
	"log"
	"net"
	"os"
)

type UnixSocket struct {
	*ipc.Server

	path string
}

func New(path string, router *ipc.Router) *UnixSocket {
	return &UnixSocket{
		Server: &ipc.Server{
			Router: router,
		},
		path: path,
	}
}

func (s *UnixSocket) ListenAndServe() error {
	if err := os.RemoveAll(s.path); err != nil {
		log.Printf("[unixsocket] remove socket error: %v", err)
		return errors.Join(err, errors.New("remove socket error"))
	}

	ln, err := net.Listen("unix", s.path)
	if err != nil {
		log.Printf("[unixsocket] listen error: %v", err)
		return errors.Join(err, errors.New("listen error"))
	}

	return s.Serve(ln)
}

func (s *UnixSocket) Register(kind int, handler ipc.Handler) {
	s.Router.Register(kind, handler)
}
