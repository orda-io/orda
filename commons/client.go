package commons

import (
	"github.com/knowhunger/ortoo/commons/context"
	"github.com/knowhunger/ortoo/commons/errors"
	"github.com/knowhunger/ortoo/commons/internal/client"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
)

//Client is a client of Ortoo which manages connections and data
type Client interface {
	Connect() error
	createDatatype()
	Close() error
	Sync() error
	SubscribeOrCreateIntCounter(key string) (chan IntCounter, chan error)
	SubscribeIntCounter(key string) (chan IntCounter, chan error)
	CreateIntCounter(key string) (chan IntCounter, chan error)
}

//NewOrtooClient creates a new Ortoo client
func NewOrtooClient(conf *OrtooClientConfig) (Client, error) {
	ctx := context.NewOrtooContext()
	cuid, err := model.NewCUID()
	if err != nil {
		return nil, log.OrtooErrorf(err, "fail to create cuid")
	}

	clientModel := &model.Client{
		CUID:       cuid,
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

type clientImpl struct {
	conf     *OrtooClientConfig
	clientID model.CUID
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
	return c.subscribeOrCreateIntCounter(key, model.StateOfDatatype_DUE_TO_CREATE)
}

func (c *clientImpl) SubscribeIntCounter(key string) (intCounterCh chan IntCounter, errCh chan error) {
	return c.subscribeOrCreateIntCounter(key, model.StateOfDatatype_DUE_TO_SUBSCRIBE)
}

func (c *clientImpl) SubscribeOrCreateIntCounter(key string) (intCounterCh chan IntCounter, errCh chan error) {
	return c.subscribeOrCreateIntCounter(key, model.StateOfDatatype_DUE_TO_SUBSCRIBE_CREATE)
}

func (c *clientImpl) subscribeOrCreateIntCounter(key string, state model.StateOfDatatype) (intCounterCh chan IntCounter, errCh chan error) {
	errCh = make(chan error)
	intCounterCh = make(chan IntCounter)

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

	ic, err := NewIntCounter(key, c.model.CUID, c.dataMgr)
	if err != nil {
		errCh <- log.OrtooErrorf(err, "fail to create intCounter")
		return
	}
	icImpl := ic.(*intCounter)
	if err := c.dataMgr.SubscribeOrCreate(icImpl, state); err != nil {
		errCh <- log.OrtooErrorf(err, "fail to subscribe intCounter")
	}

	go func() {
		if err := c.dataMgr.Sync(icImpl.GetKey()); err != nil {
			errCh <- err
		}
		intCounterCh <- icImpl
	}()

	return
}

func (c *clientImpl) Sync() error {
	return c.dataMgr.SyncAll()
}
