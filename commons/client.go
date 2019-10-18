package commons

import (
	"github.com/knowhunger/ortoo/commons/context"
	"github.com/knowhunger/ortoo/commons/errors"
	"github.com/knowhunger/ortoo/commons/internal/client"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
)

type clientImpl struct {
	conf     *OrtooClientConfig
	clientID model.Cuid
	model    *model.Client
	ctx      *context.OrtooContext
	msgMgr   *client.MessageManager
	dataMgr  *client.DataManager
}

func (c *clientImpl) Connect() error {
	if err := c.msgMgr.Connect(); err != nil {
		return log.OrtooErrorf(err, "fail to connect")
	}

	return c.msgMgr.ExchangeClientRequestResponse()
}

func (c *clientImpl) createDatatype() {

}

func (c *clientImpl) Close() error {
	if err := c.msgMgr.Close(); err != nil {
		return log.OrtooErrorf(err, "fail to close grpc connection")
	}
	return nil
}

func (c *clientImpl) CreateIntCounter(key string) (intCounterCh chan IntCounter, errCh chan error) {
	intCounterCh = make(chan IntCounter)
	errCh = make(chan error)

}

func (c *clientImpl) SubscribeOrCreateIntCounter(key string, state model.StateOfDatatype) (intCounterCh chan IntCounter, errCh chan error) {
	intCounterCh = make(chan IntCounter)
	errCh = make(chan error)

	fromDataMgr := c.dataMgr.Get(key)
	if fromDataMgr != nil {
		if fromDataMgr.GetType() == model.TypeOfDatatype_INT_COUNTER {
			log.Logger.Info("Already subscribed datatype")
			intCounterCh <- fromDataMgr.(IntCounter)
			return
		}
		errCh <- &errors.ErrSubscribeDatatype{}
		return
	}

	ic, err := NewIntCounter(key, c)
	if err != nil {
		errCh <- log.OrtooErrorf(err, "fail to create intCounter")
		return
	}
	icImpl := ic.(*intCounter)
	if err := c.dataMgr.SubscribeOrCreate(icImpl, state); err != nil {
		errCh <- log.OrtooErrorf(err, "fail to subscribe intCounter")
	}

	c.dataMgr.Sync(icImpl.GetKey())

	return
}

func (c *clientImpl) Sync() <-chan struct{} {
	//ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	//defer cancel()
	//c.serviceClient.ProcessPushPull(ctx, model.NewPushPullRequest(1))
	return nil
}

//Client is a client of Ortoo which manages connections and data
type Client interface {
	Connect() error
	createDatatype()
	Close() error
	Sync() <-chan struct{}
	SubscribeOrCreateIntCounter(key string) (chan IntCounter, chan error)
	CreateIntCounter(key string) (chan IntCounter, chan error)
}

//NewOrtooClient creates a new Ortoo client
func NewOrtooClient(conf *OrtooClientConfig) (Client, error) {
	ctx := context.NewOrtooContext()
	cuid, err := model.NewCuid()
	if err != nil {
		return nil, log.OrtooErrorf(err, "fail to create cuid")
	}

	clientModel := &model.Client{
		Cuid:       cuid,
		Alias:      conf.Alias,
		Collection: conf.CollectionName,
	}
	msgMgr := client.NewMessageManager(ctx, clientModel, conf.getServiceHost())
	dataMgr := client.NewDataManager(msgMgr)
	return &clientImpl{
		conf:     conf,
		ctx:      ctx,
		model:    clientModel,
		clientID: cuid,
		msgMgr:   msgMgr,
		dataMgr:  dataMgr,
	}, nil
}
