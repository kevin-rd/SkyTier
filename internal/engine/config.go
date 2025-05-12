package engine

import "kevin-rd/my-tier/pkg/utils"

type Config struct {
	ID        string // limit 32 bits
	VirtualIP string // e.g. "192.168.10.1/24"
	MixedPort int
	TunName   string

	Peers []string

	// Deprecated: PublicServerAddr is the address of the public server.
	PublicServerAddr string
}

func NewConfig(opts ...Option) *Config {
	c := &Config{
		ID:        utils.RandomString(16),
		MixedPort: 6780,
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

func WithVirtualIP(ip string) Option {
	return func(c *Config) {
		c.VirtualIP = ip
	}
}

func WithFixedPort(port int) Option {
	return func(c *Config) {
		if port > 0 && port < 65535 {
			c.MixedPort = port
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
