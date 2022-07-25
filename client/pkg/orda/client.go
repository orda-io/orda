package orda

import (
	gocontext "context"
	"github.com/orda-io/orda/client/pkg/context"
	errors2 "github.com/orda-io/orda/client/pkg/errors"
	"github.com/orda-io/orda/client/pkg/iface"
	"github.com/orda-io/orda/client/pkg/internal/datatypes"
	managers2 "github.com/orda-io/orda/client/pkg/internal/managers"
	"github.com/orda-io/orda/client/pkg/log"
	model2 "github.com/orda-io/orda/client/pkg/model"
	"github.com/orda-io/orda/client/pkg/types"
)

// Client is a client of Orda which manages connections and data
type Client interface {
	Connect() error
	Close() error
	Sync() error
	IsConnected() bool
	CreateDatatype(key string, typeOf model2.TypeOfDatatype, handlers *Handlers) Datatype

	CreateCounter(key string, handlers *Handlers) Counter
	SubscribeOrCreateCounter(key string, handlers *Handlers) Counter
	SubscribeCounter(key string, handlers *Handlers) Counter

	CreateMap(key string, handlers *Handlers) Map
	SubscribeOrCreateMap(key string, handlers *Handlers) Map
	SubscribeMap(key string, handlers *Handlers) Map

	CreateList(key string, handlers *Handlers) List
	SubscribeOrCreateList(key string, handlers *Handlers) List
	SubscribeList(key string, handlers *Handlers) List

	CreateDocument(key string, handlers *Handlers) Document
	SubscribeOrCreateDocument(key string, handlers *Handlers) Document
	SubscribeDocument(key string, handlers *Handlers) Document
}

type clientState uint8

const (
	notConnected clientState = iota
	connected
)

type clientImpl struct {
	state           clientState
	conf            *ClientConfig
	ctx             *context.ClientContext
	syncManager     *managers2.SyncManager
	datatypeManager *managers2.DatatypeManager
}

// NewClient creates a new Orda client
func NewClient(conf *ClientConfig, alias string) Client {
	cm := &model2.Client{
		CUID:       types.NewUID(),
		Alias:      alias,
		Collection: conf.CollectionName,
		Type:       model2.ClientType_PERSISTENT,
		SyncType:   conf.SyncType,
	}
	ctx := context.NewClientContext(gocontext.TODO(), cm)

	var syncManager *managers2.SyncManager = nil
	var datatypeManager *managers2.DatatypeManager = nil
	if conf.SyncType != model2.SyncType_LOCAL_ONLY {
		syncManager = managers2.NewSyncManager(ctx, cm, conf.ServerAddr, conf.NotificationAddr)
	}
	datatypeManager = managers2.NewDatatypeManager(ctx, syncManager)
	return &clientImpl{
		conf:            conf,
		ctx:             ctx,
		state:           notConnected,
		syncManager:     syncManager,
		datatypeManager: datatypeManager,
	}
}

func (its *clientImpl) IsConnected() bool {
	return its.state == connected
}

func (its *clientImpl) CreateDatatype(key string, typeOf model2.TypeOfDatatype, handlers *Handlers) Datatype {
	switch typeOf {
	case model2.TypeOfDatatype_COUNTER:
		return its.CreateCounter(key, handlers).(Datatype)
	case model2.TypeOfDatatype_MAP:
		return its.CreateMap(key, handlers).(Datatype)
	case model2.TypeOfDatatype_LIST:
		return its.CreateList(key, handlers).(Datatype)
	case model2.TypeOfDatatype_DOCUMENT:
		return its.CreateDocument(key, handlers).(Datatype)
	}
	return nil
}

func (its *clientImpl) Connect() (err error) {
	defer func() {
		if err == nil {
			its.state = connected
		}
	}()
	if err = its.syncManager.Connect(); err != nil {
		return errors2.ClientConnect.New(its.ctx.L(), err.Error())
	}
	err = its.syncManager.ExchangeClientRequestResponse()
	return
}

func (its *clientImpl) Close() error {
	its.state = notConnected
	its.ctx.L().Infof("close client")
	return its.syncManager.Close()
}

// methods for Document

func (its *clientImpl) CreateDocument(key string, handlers *Handlers) Document {
	return its.subscribeOrCreateDocument(key, model2.StateOfDatatype_DUE_TO_CREATE, handlers)
}

func (its *clientImpl) SubscribeDocument(key string, handlers *Handlers) Document {
	return its.subscribeOrCreateDocument(key, model2.StateOfDatatype_DUE_TO_SUBSCRIBE, handlers)
}

func (its *clientImpl) SubscribeOrCreateDocument(key string, handlers *Handlers) Document {
	return its.subscribeOrCreateDocument(key, model2.StateOfDatatype_DUE_TO_SUBSCRIBE_CREATE, handlers)
}

func (its *clientImpl) subscribeOrCreateDocument(key string, state model2.StateOfDatatype, handlers *Handlers) Document {
	datatype := its.subscribeOrCreateDatatype(key, model2.TypeOfDatatype_DOCUMENT, state, handlers)
	if datatype != nil {
		return datatype.(Document)
	}
	return nil
}

// methods for List

func (its *clientImpl) CreateList(key string, handlers *Handlers) List {
	return its.subscribeOrCreateList(key, model2.StateOfDatatype_DUE_TO_CREATE, handlers)
}

func (its *clientImpl) SubscribeList(key string, handlers *Handlers) List {
	return its.subscribeOrCreateList(key, model2.StateOfDatatype_DUE_TO_SUBSCRIBE, handlers)
}

func (its *clientImpl) SubscribeOrCreateList(key string, handlers *Handlers) List {
	return its.subscribeOrCreateList(key, model2.StateOfDatatype_DUE_TO_SUBSCRIBE_CREATE, handlers)
}

func (its *clientImpl) subscribeOrCreateList(key string, state model2.StateOfDatatype, handlers *Handlers) List {
	datatype := its.subscribeOrCreateDatatype(key, model2.TypeOfDatatype_LIST, state, handlers)
	if datatype != nil {
		return datatype.(List)
	}
	return nil
}

// methods for Map

func (its *clientImpl) CreateMap(key string, handlers *Handlers) Map {
	return its.subscribeOrCreateMap(key, model2.StateOfDatatype_DUE_TO_CREATE, handlers)
}

func (its *clientImpl) SubscribeMap(key string, handlers *Handlers) Map {
	return its.subscribeOrCreateMap(key, model2.StateOfDatatype_DUE_TO_SUBSCRIBE, handlers)
}

func (its *clientImpl) SubscribeOrCreateMap(key string, handlers *Handlers) Map {
	return its.subscribeOrCreateMap(key, model2.StateOfDatatype_DUE_TO_SUBSCRIBE_CREATE, handlers)
}

func (its *clientImpl) subscribeOrCreateMap(key string, state model2.StateOfDatatype, handlers *Handlers) Map {
	datatype := its.subscribeOrCreateDatatype(key, model2.TypeOfDatatype_MAP, state, handlers)
	if datatype != nil {
		return datatype.(Map)
	}
	return nil
}

// methods for Counter

func (its *clientImpl) CreateCounter(key string, handlers *Handlers) Counter {
	return its.subscribeOrCreateCounter(key, model2.StateOfDatatype_DUE_TO_CREATE, handlers)
}

func (its *clientImpl) SubscribeCounter(key string, handlers *Handlers) Counter {
	return its.subscribeOrCreateCounter(key, model2.StateOfDatatype_DUE_TO_SUBSCRIBE, handlers)
}

func (its *clientImpl) SubscribeOrCreateCounter(key string, handlers *Handlers) Counter {
	return its.subscribeOrCreateCounter(key, model2.StateOfDatatype_DUE_TO_SUBSCRIBE_CREATE, handlers)
}

func (its *clientImpl) subscribeOrCreateCounter(key string, state model2.StateOfDatatype, handlers *Handlers) Counter {
	datatype := its.subscribeOrCreateDatatype(key, model2.TypeOfDatatype_COUNTER, state, handlers)
	if datatype != nil {
		return datatype.(Counter)
	}
	return nil
}

func (its *clientImpl) subscribeOrCreateDatatype(
	key string,
	typeOf model2.TypeOfDatatype,
	state model2.StateOfDatatype,
	handler *Handlers,
) iface.Datatype {
	// TODO: this would be better go into datatypeManager
	if its.datatypeManager != nil {
		data, err := its.datatypeManager.ExistDatatype(key, typeOf)
		if err != nil && handler != nil {
			handler.errorHandler(nil, err)
			return nil
		}
		if data != nil {
			return data
		}
	}
	var datatype iface.Datatype
	var impl Datatype
	var errs errors2.OrdaError = &errors2.MultipleOrdaErrors{}
	var err errors2.OrdaError
	base := datatypes.NewBaseDatatype(key, typeOf, its.ctx, state)
	switch typeOf {
	case model2.TypeOfDatatype_COUNTER:
		impl, err = newCounter(base, its.datatypeManager, handler)
	case model2.TypeOfDatatype_MAP:
		impl, err = newMap(base, its.datatypeManager, handler)
	case model2.TypeOfDatatype_LIST:
		impl, err = newList(base, its.datatypeManager, handler)
	case model2.TypeOfDatatype_DOCUMENT:
		impl, err = newDocument(base, its.datatypeManager, handler)
	}
	if err != nil {
		errs = errs.Append(err)
	}
	datatype = impl.(iface.Datatype)

	if its.datatypeManager != nil {
		if err2 := its.datatypeManager.SubscribeOrCreate(datatype, state); err2 != nil {
			errs = errs.Append(err2)
		}
	}

	if handler != nil && errs.Return() != nil {
		handler.errorHandler(nil, errs.ToArray()...)
	}
	return datatype
}

func (its *clientImpl) SetLogger(logger *log.OrdaLog) {
	its.ctx.SetLogger(logger)
}

func (its *clientImpl) Sync() error {
	if its.state == connected {
		return its.datatypeManager.SyncAll()
	}
	return errors2.ClientSync.New(its.ctx.L(), "not connected")
}
