package client

import (
	"context"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
	"google.golang.org/grpc"
	"time"
)

type clientImpl struct {
	conf          *OrtooClientConfig
	clientId      model.Cuid
	conn          *grpc.ClientConn
	serviceClient model.OrtooServiceClient
	context       context.Context
	cancelFn      context.CancelFunc
	requestSeq    uint32
}

func (c *clientImpl) Connect() error {
	conn, err := grpc.Dial(c.conf.getServiceHost(), grpc.WithInsecure())
	if err != nil {
		return log.OrtooError(err, "fail to connect to Ortoo Server")
	}
	c.conn = conn
	c.serviceClient = model.NewOrtooServiceClient(c.conn)
	client := &model.Client{
		Cuid:       c.clientId,
		Alias:      c.conf.Alias,
		Collection: c.conf.CollectionName,
	}
	request := model.NewClientCreateRequest(client)
	reply, err := c.serviceClient.ClientCreate(c.context, request)
	if err != nil {
		return log.OrtooError(err, "fail to send client create")
	}

	//c.serviceClient.ClientCreate(c.context, )
	return nil
}

func (c *clientImpl) createDatatype() {

}

func (c *clientImpl) Close() error {
	if err := c.conn.Close(); err != nil {
		return log.OrtooError(err, "fail to close grpc connection")
	}
	return nil
}

func (c *clientImpl) Send() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	c.serviceClient.ProcessPushPull(ctx, model.NewPushPullRequest(1))
}

type Client interface {
	Connect() error
	createDatatype()
	Close() error
	Send()
}

func NewOrtooClient(conf *OrtooClientConfig) (Client, error) {
	ctx, cancelFn := context.WithCancel(context.TODO())
	cuid, err := model.NewCuid()
	if err != nil {
		return nil, log.OrtooError(err, "fail to create cuid")
	}
	return &clientImpl{
		conf:     conf,
		context:  ctx,
		cancelFn: cancelFn,
		clientId: cuid,
	}, nil
}
