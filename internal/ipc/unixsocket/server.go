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

func NewServer(path string) *UnixSocket {
	return &UnixSocket{
		Server: ipc.NewServer(),
		path:   path,
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
