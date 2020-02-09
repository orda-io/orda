package commons

import (
	"github.com/knowhunger/ortoo/commons/context"
	"github.com/knowhunger/ortoo/commons/errors"
	"github.com/knowhunger/ortoo/commons/internal/client"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
)

// Client is a client of Ortoo which manages connections and data
type Client interface {
	Connect() error
	Close() error
	Sync() error
	SubscribeOrCreateIntCounter(key string, handlers *IntCounterHandlers) IntCounter
	SubscribeIntCounter(key string, handlers *IntCounterHandlers) IntCounter
	CreateIntCounter(key string, handlers *IntCounterHandlers) IntCounter
}

type clientState uint8

const (
	notConnected clientState = iota
	connected
)

type clientImpl struct {
	state   clientState
	conf    *OrtooClientConfig
	model   *model.Client
	ctx     *context.OrtooContext
	msgMgr  *client.MessageManager
	dataMgr *client.DataManager
}

// NewOrtooClient creates a new Ortoo client
func NewOrtooClient(conf *OrtooClientConfig, alias string) (Client, error) {
	ctx := context.NewOrtooContext()
	cuid, err := model.NewCUID()
	if err != nil {
		return nil, log.OrtooErrorf(err, "fail to create cuid")
	}

	clientModel := &model.Client{
		CUID:       cuid,
		Alias:      alias,
		Collection: conf.CollectionName,
	}
	msgMgr := client.NewMessageManager(ctx, clientModel, conf.Address, conf.PubSubAddr)
	dataMgr := client.NewDataManager(msgMgr, clientModel.Collection, clientModel.CUID)

	return &clientImpl{
		conf:    conf,
		ctx:     ctx,
		model:   clientModel,
		state:   notConnected,
		msgMgr:  msgMgr,
		dataMgr: dataMgr,
	}, nil
}

func (c *clientImpl) Connect() (err error) {
	defer func() {
		if err != nil {
			c.state = connected
		}
	}()
	if err = c.msgMgr.Connect(); err != nil {
		return errors.NewClientError(errors.ErrClientConnect, err.Error())
	}

	err = c.msgMgr.ExchangeClientRequestResponse()
	return
}

func (c *clientImpl) Close() error {
	c.state = notConnected
	log.Logger.Infof("close client: %s", c.model.ToString())

	if err := c.msgMgr.Close(); err != nil {
		return log.OrtooErrorf(err, "fail to close grpc connection")
	}
	return nil
}

func (c *clientImpl) CreateIntCounter(key string, handlers *IntCounterHandlers) IntCounter {
	return c.subscribeOrCreateIntCounter(key, model.StateOfDatatype_DUE_TO_CREATE, handlers)
}

func (c *clientImpl) SubscribeIntCounter(key string, handlers *IntCounterHandlers) IntCounter {
	return c.subscribeOrCreateIntCounter(key, model.StateOfDatatype_DUE_TO_SUBSCRIBE, handlers)
}

func (c *clientImpl) SubscribeOrCreateIntCounter(key string, handlers *IntCounterHandlers) IntCounter {
	return c.subscribeOrCreateIntCounter(key, model.StateOfDatatype_DUE_TO_SUBSCRIBE_CREATE, handlers)
}

func (c *clientImpl) subscribeOrCreateIntCounter(key string, state model.StateOfDatatype, handlers *IntCounterHandlers) IntCounter {

	fromDataMgr := c.dataMgr.Get(key)
	if fromDataMgr != nil {
		if fromDataMgr.GetType() == model.TypeOfDatatype_INT_COUNTER {
			log.Logger.Warn("Already subscribed datatype")
			fromDataMgr.HandleSubscription()
			return nil
		} else {
			handlers.errorHandler(
				errors.NewDatatypeError(errors.ErrDatatypeSubscribe,
					"not matched type"))
			return nil
		}
	}

	ic, err := NewIntCounter(key, c.model.CUID, c.dataMgr, handlers)
	if err != nil {
		handlers.errorHandler(errors.NewDatatypeError(errors.ErrDatatypeCreate, err.Error()))
		return nil
	}
	icImpl := ic.(*intCounter)
	if err := c.dataMgr.SubscribeOrCreate(icImpl, state); err != nil {
		handlers.errorHandler(errors.NewDatatypeError(errors.ErrDatatypeSubscribe, err.Error()))
	}

	// go func() {
	// 	if c.state == connected {
	// 		if err := c.dataMgr.Sync(icImpl.GetKey()); err != nil {
	// 			icImpl.HandleError(errors.NewDatatypeError(errors.ErrDatatypeSubscribe, err.Error()))
	// 			return
	// 		}
	// 	}
	// }()
	return ic
}

func (c *clientImpl) Sync() error {
	if c.state == notConnected {
		return c.dataMgr.SyncAll()
	}
	return errors.NewClientError(errors.ErrClientNotConnected, "fail to sync")
}
