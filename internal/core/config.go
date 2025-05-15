package core

import (
	"fmt"
	"kevin-rd/my-tier/pkg/utils"
	"net"
)

type Config struct {
	ID        string // limit 32 bits
	VirtualIP string // e.g. "192.168.10.1/24"
	UDPPort   int
	TunName   string

	Peers []string

	// Deprecated: PublicServerAddr is the address of the public server.
	PublicServerAddr string
}

func NewConfig(opts ...Option) *Config {
	c := &Config{
		ID:        utils.RandomString(16),
		UDPPort:   6780,
		VirtualIP: "192.168.100.1/24",
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

type Option func(*Config)

func WithID(id string) Option {
	return func(c *Config) {
		if id == "" {
			return
		}
		c.ID = id
	}
}

func WithVirtualIP(ipStr string) Option {
	return func(c *Config) {
		ip, ipNet, err := net.ParseCIDR(ipStr)
		if err == nil && ip != nil {
			ip = ip.To4()
			if ip != nil {
				c.VirtualIP = fmt.Sprintf("%s/%d", ip.String(), ipNet.Mask)
				return
			}
		}
		ip = net.ParseIP(ipStr)
		if ip != nil {
			c.VirtualIP = fmt.Sprintf("%s/24", ip.String())
			return
		}
	}
}

func WithFixedPort(port int) Option {
	return func(c *Config) {
		if port > 0 && port < 65535 {
			c.UDPPort = port
		}
	}
}

func WithTunName(tunName string) Option {
	return func(c *Config) {
		c.TunName = tunName
	}
}

func WithPublicAddr(addr ...string) Option {
	return func(c *Config) {
		c.Peers = addr
	}
}
