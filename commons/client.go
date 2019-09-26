package commons

import (
	"github.com/knowhunger/ortoo/commons/context"
	"github.com/knowhunger/ortoo/commons/errors"
	"github.com/knowhunger/ortoo/commons/internal/client"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
)

type clientImpl struct {
	conf      *OrtooClientConfig
	clientID  model.Cuid
	model     *model.Client
	ctx       *context.OrtooContext
	reqResMgr *client.RequestResponseManager
	dataMgr   *client.DataManager
}

func (c *clientImpl) Connect() error {
	if err := c.reqResMgr.Connect(); err != nil {
		return log.OrtooError(err, "fail to connect")
	}

	return c.reqResMgr.ExchangeClientRequestResponse(c.model)
}

func (c *clientImpl) createDatatype() {

}

func (c *clientImpl) Close() error {
	if err := c.reqResMgr.Close(); err != nil {
		return log.OrtooError(err, "fail to close grpc connection")
	}
	return nil
}

func (c *clientImpl) SubscribeIntCounter(key string) (intCounterCh chan IntCounter, errCh chan error) {
	intCounterCh = make(chan IntCounter)
	errCh = make(chan error)

	fromDataMgr := c.dataMgr.Get(key)
	if fromDataMgr != nil {
		if fromDataMgr.GetType() == model.TypeOfDatatype_INT_COUNTER {
			log.Logger.Info("Already subscribed datatype")
			intCounterCh <- fromDataMgr.(IntCounter)
			return
		}
		errCh <- &errors.ErrLinkDatatype{}
		return
	}

	ic, err := NewIntCounter(key, c)
	if err != nil {
		errCh <- log.OrtooError(err, "fail to create intCounter")
		return
	}
	icImpl := ic.(*intCounter)
	if err := c.dataMgr.Subscribe(icImpl); err != nil {
		errCh <- log.OrtooError(err, "fail to subscribe intCounter")
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
	SubscribeIntCounter(key string) (chan IntCounter, chan error)
}

//NewOrtooClient creates a new Ortoo client
func NewOrtooClient(conf *OrtooClientConfig) (Client, error) {
	ctx := context.NewOrtooContext()
	cuid, err := model.NewCuid()
	if err != nil {
		return nil, log.OrtooError(err, "fail to create cuid")
	}
	reqResMgr := client.NewRequestResponseManager(ctx, conf.getServiceHost())
	dataMgr := client.NewDataManager(reqResMgr)
	model := &model.Client{
		Cuid:       cuid,
		Alias:      conf.Alias,
		Collection: conf.CollectionName,
	}
	return &clientImpl{
		conf:      conf,
		ctx:       ctx,
		model:     model,
		clientID:  cuid,
		reqResMgr: reqResMgr,
		dataMgr:   dataMgr,
	}, nil
}
