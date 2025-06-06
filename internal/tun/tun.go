// tun.go
package tun

import (
	"errors"
	"fmt"
	"github.com/songgao/water"
	"kevin-rd/my-tier/pkg/packet"
	"log"
	"sync"
)

var (
	bufPool = sync.Pool{
		New: func() any {
			return make([]byte, 1500)
		},
	}
)

type TunDevice struct {
	Iface *water.Interface
}

func NewTunDevice(name string) (*TunDevice, error) {
	config := water.Config{
		DeviceType: water.TUN,
	}
	config.Name = name

	ifce, err := water.New(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create TUN device: %v", err)
	}

	return &TunDevice{Iface: ifce}, nil
}

func (t *TunDevice) Run(outputCh chan *packet.Packet[packet.Packable]) error {
	// TUN → PeerManager
	for {
		pkt, err := t.ReadPacket()
		if err != nil {
			log.Println("tun read error:", err)
			continue
		}
		outputCh <- pkt
	}
}

func (t *TunDevice) ReadPacket() (*packet.Packet[packet.Packable], error) {
	buf := bufPool.Get().([]byte)
	defer bufPool.Put(buf)

	_, err := t.Iface.Read(buf)
	if err != nil {
		return nil, err
	}
	pkt := new(packet.Packet[packet.Packable])
	if err = pkt.Decode(buf); err != nil {
		return nil, errors.Join(packet.ErrPacketDecode, err)
	}
	return pkt, nil
}

func (t *TunDevice) WritePacket(pkt *packet.Packet[packet.Packable]) error {
	buf, err := pkt.Encode()
	if err != nil {
		return errors.Join(packet.ErrPacketEncode, err)
	}

	_, err = t.Iface.Write(buf)
	return err
}
