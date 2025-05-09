// Package engine
package engine

import (
	"fmt"
	"kevin-rd/my-tier/internal/seed"
	"kevin-rd/my-tier/internal/tun"
	"kevin-rd/my-tier/pkg/packet"
	"log"
	"net"
)

type Engine struct {
	Tun       *tun.TunDevice
	MixedAddr string // Mixed Addr local
	mux       *ServeMux

	config *Config

	seed *seed.Service
}

func New(opts ...Option) *Engine {
	cfg := NewConfig(opts...)

	e := &Engine{
		MixedAddr: "127.0.0.1:8080",
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

	if err := e.StartMixedServer(); err != nil {
		return fmt.Errorf("start mixed server fixed: %w", err)
	}
	e.mux.RegisterHandle(packet.TypeAuxPeers, e.HandleGetPeers)
	e.mux.RegisterHandle(packet.TypePing, e.HandlePing)
	log.Printf("register all hanlders to mixed server, count: %d", len(e.mux.handlers))

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
