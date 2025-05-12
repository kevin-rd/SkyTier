package main

import (
	"github.com/urfave/cli"
	"kevin-rd/my-tier/internal/engine"
	"kevin-rd/my-tier/internal/utils"
	"log"
	"os"
)

const version = "latest"

var app = &cli.App{
	Name:    "my-tier",
	Version: version,
	Usage:   "a tier network",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "id",
			Usage: "id name of this tier",
			Value: "",
		},
		cli.IntFlag{
			Name:  "fixed-port",
			Usage: "fixed port for mixed server",
			Value: 0,
		},
	},
	Action: func(c *cli.Context) {
		log.Println("starting my-tier core")
		e := engine.New(
			engine.WithID(c.String("id")),
			engine.WithVirtualIP("10.0.0.1"),
			engine.WithFixedPort(c.Int("fixed-port")),
			engine.WithTunName(""),
			engine.WithPublicAddr("127.0.0.1:6780"),
		)

		if err := e.Start(); err != nil {
			log.Fatal(err)
		}

		utils.WaitSignal([]os.Signal{os.Interrupt}, func() {
			log.Println("Stopping...")
			e.Stop()
		})
	},
}
