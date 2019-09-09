package client

import (
	"github.com/knowhunger/ortoo/commons/context"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
)

type clientImpl struct {
	conf            *OrtooClientConfig
	clientID        model.Cuid
	model           *model.Client
	ctx             *context.OrtooContext
	requestReplyMgr *requestReplyManager
}

func (c *clientImpl) Connect() error {
	err := c.requestReplyMgr.Connect()
	if err != nil {
		return log.OrtooError(err, "fail to connect")
	}

	return c.requestReplyMgr.exchangeClientRequestReply(c.model)
}

func (c *clientImpl) createDatatype() {

}

func (c *clientImpl) Close() error {
	if err := c.requestReplyMgr.Close(); err != nil {
		return log.OrtooError(err, "fail to close grpc connection")
	}
	return nil
}

func (c *clientImpl) Send() {
	//ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	//defer cancel()
	//c.serviceClient.ProcessPushPull(ctx, model.NewPushPullRequest(1))
}

//Client is a client of Ortoo which manages connections and data
type Client interface {
	Connect() error
	createDatatype()
	Close() error
	Send()
}

//NewOrtooClient creates a new Ortoo client
func NewOrtooClient(conf *OrtooClientConfig) (Client, error) {
	ctx := context.NewOrtooContext()
	cuid, err := model.NewCuid()
	if err != nil {
		return nil, log.OrtooError(err, "fail to create cuid")
	}
	requestReplyMgr := newRequestReplyManager(ctx, conf.getServiceHost())
	model := &model.Client{
		Cuid:       cuid,
		Alias:      conf.Alias,
		Collection: conf.CollectionName,
	}
	return &clientImpl{
		conf:            conf,
		ctx:             ctx,
		model:           model,
		clientID:        cuid,
		requestReplyMgr: requestReplyMgr,
	}, nil
}
