package client

import (
	"context"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
	"google.golang.org/grpc"
	"time"
)

type clientImpl struct {
	db             string
	collectionName string
	clientId       *model.Cuid
	address        string
	conn           *grpc.ClientConn
	serviceClient  model.OrtooServiceClient
}

func (c *clientImpl) Connect() error {
	conn, err := grpc.Dial(c.address, grpc.WithInsecure())
	if err != nil {
		return log.OrtooError(err, "fail to connect to Ortoo Server")
	}
	c.conn = conn
	c.serviceClient = model.NewOrtooServiceClient(c.conn)
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

func NewOrtooClient(address string) Client {
	return &clientImpl{
		address: address,
	}
}
