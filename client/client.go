package client

import (
	"github.com/knowhunger/ortoo/commons/context"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
)

type clientImpl struct {
	conf          *OrtooClientConfig
	clientId      model.Cuid
	model         *model.Client
	ctx           *context.OrtooContext
	reqRepManager *requestReplyManager
}

func (c *clientImpl) Connect() error {
	err := c.reqRepManager.Connect()
	if err != nil {
		return log.OrtooError(err, "fail to connect")
	}

	c.reqRepManager.ExchangeClientRequestReply(c.model)
	return nil
}

func (c *clientImpl) createDatatype() {

}

func (c *clientImpl) Close() error {
	if err := c.reqRepManager.Close(); err != nil {
		return log.OrtooError(err, "fail to close grpc connection")
	}
	return nil
}

func (c *clientImpl) Send() {
	//ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	//defer cancel()
	//c.serviceClient.ProcessPushPull(ctx, model.NewPushPullRequest(1))
}

type Client interface {
	Connect() error
	createDatatype()
	Close() error
	Send()
}

func NewOrtooClient(conf *OrtooClientConfig) (Client, error) {
	ctx := context.NewOrtooContext()
	cuid, err := model.NewCuid()
	if err != nil {
		return nil, log.OrtooError(err, "fail to create cuid")
	}
	NewRequestReplyManager(ctx, conf.getServiceHost())
	client := &model.Client{
		Cuid:       cuid,
		Alias:      conf.Alias,
		Collection: conf.CollectionName,
	}
	return &clientImpl{
		conf:     conf,
		ctx:      ctx,
		model:    client,
		clientId: cuid,
	}, nil
}
