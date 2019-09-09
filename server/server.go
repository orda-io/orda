package server

import (
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
	"github.com/knowhunger/ortoo/server/service"
	"google.golang.org/grpc"

	"net"
)

type OrtooServer struct {
	conf    *OrtooServerConfig
	service *service.OrtooService
	server  *grpc.Server
}

func NewOrtooServer(conf *OrtooServerConfig) *OrtooServer {
	return &OrtooServer{
		conf: conf,
	}
}

func (o *OrtooServer) Start() {
	lis, err := net.Listen("tcp", o.conf.getHostAddress())
	if err != nil {
		log.Logger.Fatalf("fail to listen: %v", err)
	}
	o.server = grpc.NewServer()
	if o.service, err = service.NewOrtooService(o.conf.Mongo); err != nil {
		panic("fail to connect MongoDB")
	}
	model.RegisterOrtooServiceServer(o.server, o.service)
	o.service.Initialize()
	if err := o.server.Serve(lis); err != nil {
		_ = log.OrtooError(err, "fail to serve grpc")
	}
	log.Logger.Info("end of start()")

}

func (o *OrtooServer) Close() {
	o.server.GracefulStop()
	log.Logger.Info("close Ortoo server")
}
