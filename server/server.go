package server

import (
	"context"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
	"github.com/knowhunger/ortoo/server/mongodb"
	"github.com/knowhunger/ortoo/server/service"
	"google.golang.org/grpc"

	"net"
)

// OrtooServer is an Ortoo server
type OrtooServer struct {
	ctx     context.Context
	conf    *OrtooServerConfig
	service *service.OrtooService
	server  *grpc.Server
	Mongo   *mongodb.RepositoryMongo
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

	lis, err := net.Listen("tcp", o.conf.getHostAddress())
	if err != nil {
		log.Logger.Fatalf("fail to listen: %v", err)
	}
	o.server = grpc.NewServer()
	if o.service, err = service.NewOrtooService(o.Mongo); err != nil {
		panic("fail to connect MongoDB")
	}
	model.RegisterOrtooServiceServer(o.server, o.service)
	err = o.service.Initialize(context.Background())
	if err != nil {
		return log.OrtooErrorf(err, "fail to initialize service")
	}
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
