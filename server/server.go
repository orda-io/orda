package main

import (
	"github.com/orda-io/orda/server/server"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "conf",
				Usage:    "server configuration file in JSON format",
				Required: true,
			},
		},

		Action: func(c *cli.Context) error {
			confFile := c.String("conf")

			conf, err := server.LoadOrdaServerConfig(confFile)
			if err != nil {
				os.Exit(1)
			}
			svr, err := server.NewOrdaServer(c.Context, conf)
			if err != nil {
				os.Exit(1)
			}
			go func() {
				if err := svr.Start(); err != nil {
					os.Exit(1)
				}
			}()
			os.Exit(svr.HandleSignals())
			return nil
		},
	}
	if err := app.Run(os.Args); err != nil {
		// TODO: should close services and resources
	}
}
