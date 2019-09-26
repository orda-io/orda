package client

import (
	"github.com/knowhunger/ortoo/commons/context"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
	"google.golang.org/grpc"
)

//RequestResponseManager is a manager exchanging request and response.
type RequestResponseManager struct {
	seq           uint32
	host          string
	conn          *grpc.ClientConn
	serviceClient model.OrtooServiceClient
	ctx           *context.OrtooContext
}

//NewRequestResponseManager ...
func NewRequestResponseManager(ctx *context.OrtooContext, host string) *RequestResponseManager {
	return &RequestResponseManager{
		seq:  0,
		host: host,
		ctx:  ctx,
	}
}

//ExchangeClientRequestResponse ...
func (r *RequestResponseManager) ExchangeClientRequestResponse(client *model.Client) error {
	request := model.NewClientRequest(client, r.seq)
	_, err := r.serviceClient.ProcessClient(r.ctx, request)
	if err != nil {
		return log.OrtooError(err, "fail to exchange clientRequestReply")
	}

	return nil

}

//Connect ...
func (r *RequestResponseManager) Connect() error {
	conn, err := grpc.Dial(r.host, grpc.WithInsecure())
	if err != nil {
		return log.OrtooError(err, "fail to connect to Ortoo Server")
	}
	r.conn = conn
	r.serviceClient = model.NewOrtooServiceClient(r.conn)
	return nil
}

//Close ...
func (r *RequestResponseManager) Close() error {
	if err := r.conn.Close(); err != nil {
		return log.OrtooError(err, "fail to close grpc connection")
	}
	return nil
}

func (r *RequestResponseManager) Sync(pppList ...*model.PushPullPack) (*model.PushPullResponse, error) {
	request := model.NewPushPullRequest(0, pppList...)
	pushPullResponse, err := r.serviceClient.ProcessPushPull(r.ctx, request)
	if err != nil {
		return nil, log.OrtooError(err, "fail to sync push pull")
	}
	return pushPullResponse, nil
}
