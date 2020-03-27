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

// OrtooServer is an Ortoo server
type OrtooServer struct {
	ctx      context.Context
	conf     *OrtooServerConfig
	service  *service.OrtooService
	server   *grpc.Server
	Mongo    *mongodb.RepositoryMongo
	notifier *notification.Notifier
}

// NewOrtooServer creates a new Ortoo server
func NewOrtooServer(ctx context.Context, conf *OrtooServerConfig) (*OrtooServer, error) {
	mongo, err := mongodb.New(ctx, conf.Mongo)
	if err != nil {
		return nil, log.OrtooError(err)
	}

	return &OrtooServer{
		ctx:   ctx,
		conf:  conf,
		Mongo: mongo,
	}, nil
}

// Start start the Ortoo Server
func (o *OrtooServer) Start() error {

	lis, err := net.Listen("tcp", o.conf.OrtooServer)
	if err != nil {
		log.Logger.Fatalf("fail to listen: %v", err)
	}
	o.notifier, err = notification.NewNotifier(o.conf.NotificationAddr)
	if err != nil {
		return log.OrtooError(err)
	}
	o.server = grpc.NewServer()
	if o.service, err = service.NewOrtooService(o.Mongo, o.notifier); err != nil {
		panic("fail to connect MongoDB")
	}
	model.RegisterOrtooServiceServer(o.server, o.service)
	fmt.Printf("%sStarted at %s\n", banner, time.Now().String())
	if err := o.server.Serve(lis); err != nil {
		_ = log.OrtooErrorf(err, "fail to serve grpc")
	}

	log.Logger.Info("end of start()")
	return nil
}

// Close closes the Ortoo server
func (o *OrtooServer) Close() {
	o.server.GracefulStop()
	log.Logger.Info("close Ortoo server")
}
