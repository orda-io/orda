package main

import (
	"context"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/server/server"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	//
	// flags := []cli.Flag{
	// 	&cli.StringFlag{Name: "conf"},
	// }

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

			conf, err := server.LoadOrtooServerConfig(confFile)
			if err != nil {
				os.Exit(1)
			}
			svr, err := server.NewOrtooServer(context.Background(), conf)
			if err != nil {
				_ = log.OrtooError(err)
				os.Exit(1)
			}
			go func() {
				if err := svr.Start(); err != nil {
					_ = log.OrtooError(err)
					os.Exit(1)
				}
			}()
			os.Exit(svr.HandleSignals())
			return nil
		},
	}
	app.Run(os.Args)
}
