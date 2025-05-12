// Package engine
package engine

import (
	"fmt"
	"kevin-rd/my-tier/internal/peer"
	"kevin-rd/my-tier/internal/router"
	"kevin-rd/my-tier/internal/seed"
	"kevin-rd/my-tier/internal/tun"
	"log"
)

type Engine struct {
	Tun *tun.TunDevice

	MixedAddr   string // Mixed Addr local
	mixedServer *MixedServer

	config *Config

	peerManager *peer.Manager
	seed        *seed.Service
}

func New(opts ...Option) *Engine {
	cfg := NewConfig(opts...)

	return &Engine{
		MixedAddr: fmt.Sprintf(":%d", cfg.MixedPort),
		config:    cfg,
	}
}

func (e *Engine) Start() error {

	if e.config.TunName != "" {
		t, err := tun.NewTunDevice(e.config.TunName)
		if err != nil {
			log.Fatal(err)
		}
		e.Tun = t
	}

	// Peers Manager
	e.peerManager = peer.NewManager(e.config.ID, e.config.VirtualIP, e.config.Peers...)

	// ListenAndServe Mixed Server
	e.mixedServer = &MixedServer{
		ListenAddr: e.MixedAddr,
		router:     router.NewRouter(e.Tun, e.peerManager),
	}
	log.Printf("start mixed server on: %v", e.mixedServer.ListenAddr)
	if err := e.mixedServer.ListenAndServe(); err != nil {
		return fmt.Errorf("start mixed server error: %w", err)
	}

	return nil
}

func (e *Engine) Stop() {

}
