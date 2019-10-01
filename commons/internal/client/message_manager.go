package client

import (
	"github.com/knowhunger/ortoo/commons/context"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
	"google.golang.org/grpc"
)

//MessageManager is a manager exchanging request and response.
type MessageManager struct {
	seq           uint32
	host          string
	client        *model.Client
	conn          *grpc.ClientConn
	serviceClient model.OrtooServiceClient
	ctx           *context.OrtooContext
}

//NewMessageManager ...
func NewMessageManager(ctx *context.OrtooContext, client *model.Client, host string) *MessageManager {
	return &MessageManager{
		seq:    0,
		ctx:    ctx,
		host:   host,
		client: client,
	}
}

//ExchangeClientRequestResponse ...
func (r *MessageManager) ExchangeClientRequestResponse() error {
	request := model.NewClientRequest(r.NextSeq(), r.client)
	_, err := r.serviceClient.ProcessClient(r.ctx, request)
	if err != nil {
		return log.OrtooErrorf(err, "fail to exchange clientRequestReply")
	}

	return nil

}

func (r *MessageManager) NextSeq() uint32 {
	currentSeq := r.seq
	r.seq++
	return currentSeq
}

//Connect ...
func (r *MessageManager) Connect() error {
	conn, err := grpc.Dial(r.host, grpc.WithInsecure())
	if err != nil {
		return log.OrtooErrorf(err, "fail to connect to Ortoo Server")
	}
	r.conn = conn
	r.serviceClient = model.NewOrtooServiceClient(r.conn)
	return nil
}

//Close ...
func (r *MessageManager) Close() error {
	if err := r.conn.Close(); err != nil {
		return log.OrtooErrorf(err, "fail to close grpc connection")
	}
	return nil
}

func (r *MessageManager) Sync(pppList ...*model.PushPullPack) (*model.PushPullResponse, error) {
	request := model.NewPushPullRequest(r.NextSeq(), r.client, pppList...)
	pushPullResponse, err := r.serviceClient.ProcessPushPull(r.ctx, request)
	if err != nil {
		return nil, log.OrtooErrorf(err, "fail to sync push pull")
	}
	return pushPullResponse, nil
}
