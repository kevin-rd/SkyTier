package main

import (
	"bytes"
	"fmt"
	"github.com/urfave/cli/v2"
	"kevin-rd/my-tier/pkg/packet"
	"net"
	"time"
)

const version = "latest"

var app = &cli.App{
	Name:    "skytier-cli",
	Version: version,
	Usage:   "A simple, decentralized mesh VPN.",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "tun-name",
			Usage: "tun name",
		},
	},
	Action: func(c *cli.Context) error {
		return nil
	},
	Commands: []*cli.Command{
		{
			Name:  "test",
			Usage: "test tier network",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "host", Usage: "server host", Value: "127.0.0.1"},
				&cli.IntFlag{Name: "port", Usage: "server port", Value: 6780},
				&cli.UintFlag{Name: "type", Usage: "Packet type", Value: packet.TypePing},
				&cli.StringFlag{Name: "data", Usage: "Packet body data", Value: "hello"},
			},
			Action: func(c *cli.Context) error {
				addr := fmt.Sprintf("%s:%d", c.String("host"), c.Int("port"))
				conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
				if err != nil {
					return fmt.Errorf("failed to connect: %w", err)
				}
				defer conn.Close()

				pkt := &packet.Packet{
					Version: packet.ProtocolVersion,
					Type:    byte(c.Uint("type")),
					Length:  uint16(len(c.String("data"))),
					Payload: []byte(c.String("data")),
				}
				data, err := pkt.Encode()
				if err != nil {
					return fmt.Errorf("encode error: %w", err)
				}

				if _, err := conn.Write(data); err != nil {
					return fmt.Errorf("send error: %w", err)
				}
				fmt.Println("✅ Sent packet.")

				// Read response
				buf := make([]byte, 4096)
				n, err := conn.Read(buf)
				if err != nil {
					return fmt.Errorf("read error: %w", err)
				}

				resp, err := packet.ReadPacket(bytes.NewReader(buf[:n]))
				if err != nil {
					return fmt.Errorf("decode error: %w", err)
				}
				fmt.Println("✅ Received packet:", resp.Payload)

				for {

				}
				return nil
			},
		},
	},
}
