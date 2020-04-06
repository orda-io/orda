package server

import (
	"context"
	"fmt"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/ortoo/model"
	"github.com/knowhunger/ortoo/server/mongodb"
	"github.com/knowhunger/ortoo/server/notification"
	"github.com/knowhunger/ortoo/server/service"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const banner = `
▒█████   ██▀███  ▄▄▄█████▓ ▒█████   ▒█████
▒██▒  ██▒▓██ ▒ ██▒▓  ██▒ ▓▒▒██▒  ██▒▒██▒  ██▒
▒██░  ██▒▓██ ░▄█ ▒▒ ▓██░ ▒░▒██░  ██▒▒██░  ██▒
▒██   ██░▒██▀▀█▄  ░ ▓██▓ ░ ▒██   ██░▒██   ██░
░ ████▓▒░░██▓ ▒██▒  ▒██▒ ░ ░ ████▓▒░░ ████▓▒░
░ ▒░▒░▒░ ░ ▒▓ ░▒▓░  ▒ ░░   ░ ▒░▒░▒░ ░ ▒░▒░▒░
░ ▒ ▒░   ░▒ ░ ▒░    ░      ░ ▒ ▒░   ░ ▒ ▒░
░ ░ ░ ▒    ░░   ░   ░      ░ ░ ░ ▒  ░ ░ ░ ▒
░ ░     ░                  ░ ░      ░ ░
`

const defaultGracefulTimeout = 10 * time.Second

// OrtooServer is an Ortoo server
type OrtooServer struct {
	ctx      context.Context
	conf     *OrtooServerConfig
	service  *service.OrtooService
	server   *grpc.Server
	Mongo    *mongodb.RepositoryMongo
	notifier *notification.Notifier
	closed   bool
	closedCh chan struct{}
}

// NewOrtooServer creates a new Ortoo server
func NewOrtooServer(ctx context.Context, conf *OrtooServerConfig) (*OrtooServer, error) {
	mongo, err := mongodb.New(ctx, conf.Mongo)
	if err != nil {
		return nil, log.OrtooError(err)
	}

	return &OrtooServer{
		ctx:      ctx,
		conf:     conf,
		Mongo:    mongo,
		closedCh: make(chan struct{}),
	}, nil
}

// Start start the Ortoo Server
func (its *OrtooServer) Start() error {

	lis, err := net.Listen("tcp", its.conf.OrtooServer)
	if err != nil {
		log.Logger.Fatalf("fail to listen: %v", err)
	}
	its.notifier, err = notification.NewNotifier(its.conf.Notification)
	if err != nil {
		return log.OrtooError(err)
	}
	its.server = grpc.NewServer()
	if its.service, err = service.NewOrtooService(its.Mongo, its.notifier); err != nil {
		panic("fail to connect MongoDB")
	}
	model.RegisterOrtooServiceServer(its.server, its.service)
	fmt.Printf("%sStarted at %s\n", banner, time.Now().String())
	if err := its.server.Serve(lis); err != nil {
		_ = log.OrtooErrorf(err, "fail to serve grpc")
	}

	log.Logger.Info("end of start()")
	return nil
}

// Close closes the Ortoo server
func (its *OrtooServer) Close() {
	its.server.GracefulStop()
	log.Logger.Info("close Ortoo server")
}

func (its *OrtooServer) Shutdown(graceful bool) error {
	if graceful {
		log.Logger.Infof("Gracefully shutdown server")
		its.server.GracefulStop()
	} else {
		log.Logger.Infof("Stop server")
		its.server.Stop()
	}
	return nil
}

func (its *OrtooServer) HandleSignals() int {
	log.Logger.Infof("Handle signals")
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

	var sig os.Signal
	select {
	case s := <-signalCh:
		sig = s
	case <-its.closedCh:
		return 0
	}

	graceful := false
	if sig == syscall.SIGINT || sig == syscall.SIGTERM {
		graceful = true
	}

	gracefulCh := make(chan struct{})
	go func() {
		if err := its.Shutdown(graceful); err != nil {
			return
		}
		close(gracefulCh)
	}()

	select {
	case <-signalCh:
		return 1
	case <-time.After(defaultGracefulTimeout):
		return 1
	case <-gracefulCh:
		return 0
	}
}
