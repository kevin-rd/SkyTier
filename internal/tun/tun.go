// tun.go
package tun

import (
	"fmt"
	"github.com/songgao/water"
)

type TunDevice struct {
	Ifce *water.Interface
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

	return &TunDevice{Ifce: ifce}, nil
}

func (t *TunDevice) ReadPacket() ([]byte, error) {
	packet := make([]byte, 1500)
	n, err := t.Ifce.Read(packet)
	if err != nil {
		return nil, err
	}
	return packet[:n], nil
}

func (t *TunDevice) WritePacket(packet []byte) error {
	_, err := t.Ifce.Write(packet)
	return err
}
