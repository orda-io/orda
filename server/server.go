package main

import (
	"context"
	"flag"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/server/server"
	"os"
)

func main() {
	confFile := flag.String("conf", "", "configuration file path")
	flag.Parse()

	conf, err := server.LoadOrtooServerConfig(*confFile)
	if err != nil {
		log.Logger.Errorf("fail to load server config: %s", *confFile)
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
}
