package ortoo

import (
	"fmt"
	"github.com/knowhunger/ortoo/ortoo/context"
	"github.com/knowhunger/ortoo/ortoo/errors"
	"github.com/knowhunger/ortoo/ortoo/internal/managers"
	"github.com/knowhunger/ortoo/ortoo/model"
)

// Client is a client of Ortoo which manages connections and data
type Client interface {
	Connect() error
	Close() error
	Sync() error
	IsConnected() bool
	CreateDatatype(key string, typeOf model.TypeOfDatatype) model.Datatype

	CreateIntCounter(key string, handlers *Handlers) IntCounter
	SubscribeOrCreateIntCounter(key string, handlers *Handlers) IntCounter
	SubscribeIntCounter(key string, handlers *Handlers) IntCounter

	CreateHashMap(key string, handlers *Handlers) HashMap
}

type clientState uint8

const (
	notConnected clientState = iota
	connected
)

type clientImpl struct {
	state           clientState
	model           *model.Client
	conf            *ClientConfig
	ctx             *context.OrtooContext
	messageManager  *managers.MessageManager
	datatypeManager *managers.DatatypeManager
}

// NewClient creates a new Ortoo client
func NewClient(conf *ClientConfig, alias string) Client {
	ctx := context.NewOrtooContext(alias)
	clientModel := &model.Client{
		CUID:       model.NewCUID(),
		Alias:      alias,
		Collection: conf.CollectionName,
		SyncType:   conf.SyncType,
	}

	var notificationManager *managers.NotificationManager
	switch conf.SyncType {
	case model.SyncType_LOCAL_ONLY, model.SyncType_MANUALLY:
		notificationManager = nil
	case model.SyncType_NOTIFIABLE:
		notificationManager = managers.NewNotificationManager(ctx, conf.NotificationAddr)
	}
	var messageManager *managers.MessageManager = nil
	var datatypeManager *managers.DatatypeManager = nil
	if conf.SyncType != model.SyncType_LOCAL_ONLY {
		messageManager = managers.NewMessageManager(ctx, clientModel, conf.Address, notificationManager)
		datatypeManager = managers.NewDatatypeManager(ctx, messageManager, notificationManager, clientModel.Collection, clientModel.CUID)
	}
	return &clientImpl{
		conf:            conf,
		ctx:             ctx,
		model:           clientModel,
		state:           notConnected,
		messageManager:  messageManager,
		datatypeManager: datatypeManager,
	}
}

func (c *clientImpl) IsConnected() bool {
	return c.state == connected
}

func (c *clientImpl) CreateDatatype(key string, typeOf model.TypeOfDatatype) model.Datatype {
	switch typeOf {
	case model.TypeOfDatatype_INT_COUNTER:
		return c.CreateIntCounter(key, nil).(model.Datatype)
	case model.TypeOfDatatype_HASH_MAP:
		return c.CreateHashMap(key, nil).(model.Datatype)
	}
	return nil
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

func (c *clientImpl) CreateHashMap(key string, handlers *Handlers) HashMap {
	return c.subscribeOrCreateHashMap(key, model.StateOfDatatype_DUE_TO_CREATE, handlers)
}

func (c *clientImpl) SubscribeHashMap(key string, handlers *Handlers) HashMap {
	return c.subscribeOrCreateHashMap(key, model.StateOfDatatype_DUE_TO_SUBSCRIBE, handlers)
}

func (c *clientImpl) SubscribeOrCreateHashMap(key string, handlers *Handlers) HashMap {
	return c.subscribeOrCreateHashMap(key, model.StateOfDatatype_DUE_TO_SUBSCRIBE_CREATE, handlers)
}

func (c *clientImpl) subscribeOrCreateHashMap(key string, state model.StateOfDatatype, handlers *Handlers) HashMap {
	datatype := c.subscribeOrCreateDatatype(key, model.TypeOfDatatype_HASH_MAP, state, handlers)
	if datatype != nil {
		return datatype.(HashMap)
	}
	return nil
}

func (c *clientImpl) CreateIntCounter(key string, handlers *Handlers) IntCounter {
	return c.subscribeOrCreateIntCounter(key, model.StateOfDatatype_DUE_TO_CREATE, handlers)
}

func (c *clientImpl) SubscribeIntCounter(key string, handlers *Handlers) IntCounter {
	return c.subscribeOrCreateIntCounter(key, model.StateOfDatatype_DUE_TO_SUBSCRIBE, handlers)
}

func (c *clientImpl) SubscribeOrCreateIntCounter(key string, handlers *Handlers) IntCounter {
	return c.subscribeOrCreateIntCounter(key, model.StateOfDatatype_DUE_TO_SUBSCRIBE_CREATE, handlers)
}

func (c *clientImpl) subscribeOrCreateIntCounter(key string, state model.StateOfDatatype, handlers *Handlers) IntCounter {
	datatype := c.subscribeOrCreateDatatype(key, model.TypeOfDatatype_INT_COUNTER, state, handlers)
	if datatype != nil {
		return datatype.(IntCounter)
	}
	return nil
}

func (c *clientImpl) subscribeOrCreateDatatype(
	key string,
	typeOf model.TypeOfDatatype,
	state model.StateOfDatatype,
	handler *Handlers,
) model.Datatype {
	if c.datatypeManager != nil {
		datatypeFromDM := c.datatypeManager.Get(key)
		if datatypeFromDM != nil {
			if datatypeFromDM.GetType() == typeOf {
				c.ctx.Logger.Warnf("already subscribed datatype '%s'", key)
				return datatypeFromDM
			}
			err := errors.NewDatatypeError(errors.ErrDatatypeSubscribe,
				fmt.Sprintf("not matched type: %s vs %s", typeOf.String(), datatypeFromDM.GetType().String()))
			if handler != nil {
				handler.errorHandler(nil, err)
			}
		}
	}
	var datatype model.Datatype
	var impl interface{}

	switch typeOf {
	case model.TypeOfDatatype_INT_COUNTER:
		impl = newIntCounter(key, c.model.CUID, c.datatypeManager, handler)
	case model.TypeOfDatatype_HASH_MAP:
		impl = newHashMap(key, c.model.CUID, c.datatypeManager, handler)
	}
	datatype = impl.(model.Datatype)

	if c.datatypeManager != nil {
		if err := c.datatypeManager.SubscribeOrCreate(datatype, state); err != nil {
			err := errors.NewDatatypeError(errors.ErrDatatypeSubscribe, err.Error())
			if handler != nil {
				handler.errorHandler(nil, err)
			}
		}
	}
	return datatype
}

func (c *clientImpl) Sync() error {
	if c.state == notConnected {
		return c.datatypeManager.SyncAll()
	}
	return errors.NewClientError(errors.ErrClientNotConnected, "fail to sync")
}
