package engine

type Config struct {
	VirtualIP string // e.g. "192.168.10.1/24"
	MixedPort int
	TunName   string

	// PublicAddr is the address of the public server.
	PublicServerAddr string
}

type Option func(*Config)

func WithVirtualIP(ip string) Option {
	return func(c *Config) {
		c.VirtualIP = ip
	}
}

func WithFixedPort(port int) Option {
	return func(c *Config) {
		c.MixedPort = port
	}
}

func WithTunName(tunName string) Option {
	return func(c *Config) {
		c.TunName = tunName
	}
}

func WithPublicAddr(addr string) Option {
	return func(c *Config) {
		c.PublicServerAddr = addr
	}
}

func NewConfig(opts ...Option) *Config {
	c := &Config{}
	for _, opt := range opts {
		opt(c)
	}
	return c
}
