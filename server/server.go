package server

import (
	"context"
	"github.com/golang/protobuf/proto"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
	"google.golang.org/grpc"

	"net"
)

type OrtooServer struct {
	conf   *OrtooConfig
	server *grpc.Server
}

func (o *OrtooServer) ProcessPushPull(ctx context.Context, in *model.PushPullRequest) (*model.PushPullReply, error) {
	log.Logger.Infof("Received: %v", proto.MarshalTextString(in))
	return &model.PushPullReply{Id: in.Id}, nil
}

func NewOrtooServer(conf *OrtooConfig) *OrtooServer {
	return &OrtooServer{
		conf: conf,
	}
}

func (o *OrtooServer) Start() {
	lis, err := net.Listen("tcp", o.conf.getHostAddress())
	if err != nil {
		log.Logger.Fatalf("failed to listen: %v", err)
	}
	o.server = grpc.NewServer()
	model.RegisterOrtooServiceServer(o.server, o)
	if err := o.server.Serve(lis); err != nil {
		_ = log.OrtooError(err, "failed to serve")
	}

	log.Logger.Info("end of start()")
}

func (o *OrtooServer) Close() {
	o.server.GracefulStop()
	log.Logger.Info("close Ortoo server")
}
