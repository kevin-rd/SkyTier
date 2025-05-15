// Package engine
package core

import (
	"fmt"
	"kevin-rd/my-tier/internal/ipc/unixsocket"
	"kevin-rd/my-tier/internal/peer"
	"kevin-rd/my-tier/internal/router"
	"kevin-rd/my-tier/internal/seed"
	"kevin-rd/my-tier/internal/tun"
	"kevin-rd/my-tier/pkg/ipc/message"
	ipc_unix "kevin-rd/my-tier/pkg/ipc/unix_socket"
	"log"
	"sync"
)

type Core struct {
	config *Config

	Tun *tun.TunDevice

	// unix socket server
	UnixSocket *unix_socket.UnixSocket

	// udp server
	udpServer *UDPServer

	peerManager *peer.Manager
	seed        *seed.Service
}

func New(opts ...Option) *Core {
	cfg := NewConfig(opts...)

	return &Core{
		config: cfg,
	}
}

func (c *Core) Run() error {

	if c.config.TunName != "" {
		t, err := tun.NewTunDevice(c.config.TunName)
		if err != nil {
			log.Fatal(err)
		}
		c.Tun = t
	}

	wg := &sync.WaitGroup{}
	wg.Add(3)

	// Peers Manager
	c.peerManager = peer.NewManager(c.config.ID, c.config.VirtualIP, c.config.Peers...)
	go func() {
		defer wg.Done()

		if err := c.peerManager.Manage(); err != nil {
			log.Fatalf("[core] peers manager error: %v", err)
		}
	}()

	// Unix Socket Server
	c.UnixSocket = unix_socket.NewServer(ipc_unix.UNIX_SOCKET_PATH)
	c.UnixSocket.Register(message.KindPeers, c.UnixSocket.HandleGetPeers(c.peerManager.GetPeers))
	log.Printf("[core] start unix socket server on: %v", ipc_unix.UNIX_SOCKET_PATH)
	go func() {
		defer wg.Done()
		if err := c.UnixSocket.ListenAndServe(); err != nil {
			log.Fatalf("[core] start unix socket server error: %v", err)
		}
	}()

	// UDP Server
	c.udpServer = &UDPServer{
		ListenAddr: fmt.Sprintf(":%d", c.config.UDPPort),
		router:     router.NewRouter(c.Tun, c.peerManager),
	}
	log.Printf("[core] start udp server on: %v", c.udpServer.ListenAddr)
	go func() {
		defer wg.Done()
		if err := c.udpServer.ListenAndServe(); err != nil {
			log.Fatalf("[core] start udp server error: %v", err)
		}
	}()

	wg.Wait()
	log.Println("[core] all server done.")
	return nil
}

func (c *Core) Stop() {

}
