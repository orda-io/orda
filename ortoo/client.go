package ortoo

import (
	"github.com/knowhunger/ortoo/ortoo/context"
	"github.com/knowhunger/ortoo/ortoo/errors"
	"github.com/knowhunger/ortoo/ortoo/internal/managers"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/ortoo/model"
)

// Client is a client of Ortoo which manages connections and data
type Client interface {
	Connect() error
	Close() error
	Sync() error
	IsConnected() bool
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
	state           clientState
	model           *model.Client
	conf            *OrtooClientConfig
	ctx             *context.OrtooContext
	messageManager  *managers.MessageManager
	datatypeManager *managers.DatatypeManager
}

// NewOrtooClient creates a new Ortoo client
func NewOrtooClient(conf *OrtooClientConfig, alias string) (Client, error) {
	ctx := context.NewOrtooContext(alias)
	cuid, err := model.NewCUID()
	if err != nil {
		return nil, log.OrtooErrorf(err, "fail to create cuid")
	}

	clientModel := &model.Client{
		CUID:       cuid,
		Alias:      alias,
		Collection: conf.CollectionName,
		SyncType:   conf.SyncType,
	}
	var notificationManager *managers.NotificationManager
	switch conf.SyncType {
	case model.SyncType_MANUALLY:
		notificationManager = nil
	case model.SyncType_NOTIFIABLE:
		notificationManager = managers.NewNotificationManager(ctx, conf.NotificationAddr)
	}

	messageManager := managers.NewMessageManager(ctx, clientModel, conf.Address, notificationManager)
	datatypeManager := managers.NewDatatypeManager(ctx, messageManager, notificationManager, clientModel.Collection, clientModel.CUID)

	return &clientImpl{
		conf:            conf,
		ctx:             ctx,
		model:           clientModel,
		state:           notConnected,
		messageManager:  messageManager,
		datatypeManager: datatypeManager,
	}, nil
}

func (c *clientImpl) IsConnected() bool {
	return c.state == connected
}

func (c *clientImpl) Connect() (err error) {
	defer func() {
		if err != nil {
			c.state = connected
		}
	}()
	if err = c.messageManager.Connect(); err != nil {
		return errors.NewClientError(errors.ErrClientConnect, err.Error())
	}

	err = c.messageManager.ExchangeClientRequestResponse()
	return
}

func (c *clientImpl) Close() error {
	c.state = notConnected
	c.ctx.Logger.Infof("close client: %s", c.model.ToString())

	return c.messageManager.Close()
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
	datatypeFromDM := c.datatypeManager.Get(key)
	if datatypeFromDM != nil {
		if datatypeFromDM.GetType() == model.TypeOfDatatype_INT_COUNTER {
			c.ctx.Logger.Warn("Already subscribed datatype")
			// datatypeFromDM.HandleStateChange()
			return datatypeFromDM.(*intCounter)
		}
		handlers.errorHandler(errors.NewDatatypeError(errors.ErrDatatypeSubscribe, "not matched type"))
		return nil
	}

	ic, err := NewIntCounter(key, c.model.CUID, c.datatypeManager, handlers)
	if err != nil {
		handlers.errorHandler(errors.NewDatatypeError(errors.ErrDatatypeCreate, err.Error()))
		return nil
	}
	icImpl := ic.(*intCounter)
	if err := c.datatypeManager.SubscribeOrCreate(icImpl, state); err != nil {
		handlers.errorHandler(errors.NewDatatypeError(errors.ErrDatatypeSubscribe, err.Error()))
	}
	return ic
}

func (c *clientImpl) Sync() error {
	if c.state == notConnected {
		return c.datatypeManager.SyncAll()
	}
	return errors.NewClientError(errors.ErrClientNotConnected, "fail to sync")
}
