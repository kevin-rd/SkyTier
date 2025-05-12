package main

import (
	cli "github.com/urfave/cli/v2"
	"kevin-rd/my-tier/internal/engine"
	"kevin-rd/my-tier/pkg/utils"
	"log"
	"os"
)

const version = "latest"

var app = &cli.App{
	Name:      "skytier-core",
	Version:   version,
	Usage:     "A simple, decentralized mesh VPN.",
	UsageText: "skytier-core [global options]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "id",
			Usage: "id name of this tier",
			Value: "",
		},
		&cli.IntFlag{
			Name:  "fixed-port",
			Usage: "fixed port for mixed server",
			Value: 0,
		},
		&cli.StringSliceFlag{
			Name:    "peer",
			Aliases: []string{"p"},
			Usage:   "remote peer addr",
			Value:   nil,
		},
	},
	Action: func(c *cli.Context) error {
		log.Println("starting my-tier core")
		e := engine.New(
			engine.WithID(c.String("id")),
			engine.WithVirtualIP("10.0.0.1"),
			engine.WithFixedPort(c.Int("fixed-port")),
			engine.WithTunName(""),
			engine.WithPublicAddr(c.StringSlice("peer")...),
		)

		if err := e.Start(); err != nil {
			log.Fatal(err)
		}

		utils.WaitSignal([]os.Signal{os.Interrupt}, func() {
			log.Println("Stopping...")
			e.Stop()
		})
		return nil
	},
}
