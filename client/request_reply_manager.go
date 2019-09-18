package client

import (
	"github.com/knowhunger/ortoo/commons/context"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
	"google.golang.org/grpc"
)

type requestResponseManager struct {
	seq           uint32
	host          string
	conn          *grpc.ClientConn
	serviceClient model.OrtooServiceClient
	ctx           *context.OrtooContext
}

func newRequestResponseManager(ctx *context.OrtooContext, host string) *requestResponseManager {
	return &requestResponseManager{
		seq:  0,
		host: host,
		ctx:  ctx,
	}
}

func (r *requestResponseManager) exchangeClientRequestResponse(client *model.Client) error {
	request := model.NewClientRequest(client, r.seq)
	_, err := r.serviceClient.ProcessClient(r.ctx, request)
	if err != nil {
		return log.OrtooError(err, "fail to exchange clientRequestReply")
	}

	return nil

}

func (r *requestResponseManager) Connect() error {
	conn, err := grpc.Dial(r.host, grpc.WithInsecure())
	if err != nil {
		return log.OrtooError(err, "fail to connect to Ortoo Server")
	}
	r.conn = conn
	r.serviceClient = model.NewOrtooServiceClient(r.conn)
	return nil
}

func (r *requestResponseManager) Close() error {
	if err := r.conn.Close(); err != nil {
		return log.OrtooError(err, "fail to close grpc connection")
	}
	return nil
}
