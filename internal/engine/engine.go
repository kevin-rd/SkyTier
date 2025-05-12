// Package engine
package engine

import (
	"fmt"
	"kevin-rd/my-tier/internal/peer"
	"kevin-rd/my-tier/internal/seed"
	"kevin-rd/my-tier/internal/tun"
	"kevin-rd/my-tier/pkg/mixed_server"
	"kevin-rd/my-tier/pkg/packet"
	"log"
	"net"
)

type Engine struct {
	Tun *tun.TunDevice

	MixedAddr   string // Mixed Addr local
	mixedServer *mixed_server.MixedServer

	config *Config

	peerManager *peer.Manager
	seed        *seed.Service
}

func New(opts ...Option) *Engine {
	cfg := NewConfig(opts...)

	e := &Engine{
		MixedAddr: fmt.Sprintf(":%d", cfg.MixedPort),
		config:    cfg,
	}

	if cfg.TunName != "" {
		t, err := tun.NewTunDevice(cfg.TunName)
		if err != nil {
			log.Fatal(err)
		}
		e.Tun = t
	}

	return e
}

func (e *Engine) Start() error {

	// Peers Manager
	e.peerManager = peer.NewManager(e.config.ID, e.config.VirtualIP, e.config.Peers...)

	// ListenAndServe Mixed Server
	router := mixed_server.NewServeMux()
	e.mixedServer = &mixed_server.MixedServer{
		ListenAddr: e.MixedAddr,
		Handler:    router,
	}
	router.RegisterHandle(packet.TypeAuxPeers, e.HandleGetPeers)
	router.RegisterHandle(packet.TypePing, e.HandlePing)
	router.RegisterHandle(packet.TypeHandshake, e.HandleHandshake)
	log.Printf("register all hanlders to mixed server")
	log.Printf("start mixed server on: %v", e.mixedServer.ListenAddr)
	if err := e.mixedServer.ListenAndServe(); err != nil {
		return fmt.Errorf("start mixed server error: %w", err)
	}

	return nil
}

func (e *Engine) Stop() {

}

// deprecated: conn 连接远端peers
func (e *Engine) conn() error {
	// 监听端口
	conn, err := net.Dial("udp", e.MixedAddr)
	if err != nil {
		return err
	}

	defer conn.Close()
	// 从对端读取数据 -> 写入 TUN
	go func() {
		buf := make([]byte, 1500)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				log.Println("peer read error:", err)
				return
			}
			err = e.Tun.WritePacket(buf[:n])
			if err != nil {
				log.Println("TUN write error:", err)
			}
		}
	}()

	//// 从 TUN 读取数据 -> 发送给对端
	//for {
	//	packet, err := e.Tun.ReadPacket()
	//	if err != nil {
	//		log.Println("TUN read error:", err)
	//		return err
	//	}
	//	_, err = conn.Write(packet)
	//	if err != nil {
	//		log.Println("peer write error:", err)
	//		return err
	//	}
	//}
	return nil
}
