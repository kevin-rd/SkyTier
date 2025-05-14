package main

import (
	"bytes"
	"fmt"
	"github.com/urfave/cli/v2"
	"kevin-rd/my-tier/pkg/ipc/message"
	"kevin-rd/my-tier/pkg/ipc/unix_socket"
	"kevin-rd/my-tier/pkg/packet"
	"log"
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
		subTest,
		subPeers,
	},
}

var subTest = &cli.Command{
	Name:  "test",
	Usage: "test tier network",
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "host", Usage: "server host", Value: "127.0.0.1"},
		&cli.IntFlag{Name: "port", Usage: "server port", Value: 6780},
		&cli.UintFlag{Name: "type", Usage: "Packet type", Value: uint(packet.TypeHandshake)},
		&cli.StringFlag{Name: "data", Usage: "Packet body data", Value: "hello"},
	},
	Action: func(c *cli.Context) error {
		addr := fmt.Sprintf("%s:%d", c.String("host"), c.Int("port"))
		conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
		if err != nil {
			return fmt.Errorf("failed to connect: %w", err)
		}
		defer func(conn net.Conn) {
			_ = conn.Close()
		}(conn)

		pkt := &packet.Packet[packet.Packable]{
			Version: packet.ProtocolVersion,
			Type:    byte(c.Uint("type")),
			Length:  uint16(len(c.String("data"))),
			Payload: &packet.PayloadHandshake{
				ID:     [32]byte{'h', 'e', 'l', 'l', 'o'},
				DHCP:   false,
				IpCidr: "192.168.5.3/24",
			},
		}

		writer := packet.NewWriter(conn)
		if _, err := writer.WriteP(pkt); err != nil {
			return fmt.Errorf("send error: %w", err)
		}

		fmt.Println("✅ Sent packet.")

		// Read response
		buf := make([]byte, 4096)
		n, err := conn.Read(buf)
		if err != nil {
			return fmt.Errorf("read error: %w", err)
		}

		resp, err := packet.ReadPacketOnce(bytes.NewReader(buf[:n]))
		if err != nil {
			return fmt.Errorf("decode error: %w", err)
		}
		fmt.Printf("✅ Received packet: %v", resp.Payload)

		return nil
	},
}

var subPeers = &cli.Command{
	Name:  "peers",
	Usage: "Get peers",
	Action: func(c *cli.Context) error {
		resp, err := unix_socket.Get(&message.Message{
			Kind: message.KindPeers,
		})
		if err != nil {
			log.Fatalf("get peers error: %v", err)
		}
		// todo: 显示
		log.Println(resp)
		return nil
	},
}
