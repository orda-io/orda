package context

import (
	"context"
	"github.com/orda-io/orda/client/pkg/constants"
	"github.com/orda-io/orda/client/pkg/iface"
	"github.com/orda-io/orda/client/pkg/log"
	"github.com/orda-io/orda/client/pkg/model"
)

// OrdaContext is a context used in Orda
type OrdaContext interface {
	context.Context
	L() *log.OrdaLog
	Ctx() context.Context
	SetLogger(l *log.OrdaLog)
	UpdateTags(tag1, tag2 string)
}

type ordaContext struct {
	context.Context
	logger *log.OrdaLog
}

func (its *ordaContext) UpdateTags(tag1, tag2 string) {
	its.logger.SetTags(tag1, tag2)
}

type ClientContext struct {
	OrdaContext
	Client *model.Client
}

type DatatypeContext struct {
	*ClientContext
	Data iface.BaseDatatype
}

func NewOrdaContext(ctx context.Context, tag1 string, tag2 string) OrdaContext {
	logger := log.NewWithTags(tag1, tag2)
	return &ordaContext{
		Context: ctx,
		logger:  logger,
	}
}

func NewClientContext(ctx context.Context, client *model.Client) *ClientContext {
	return &ClientContext{
		OrdaContext: NewOrdaContext(ctx, constants.TagSdkClient, client.GetSummary()),
		Client:      client,
	}
}

func NewDatatypeContext(clientContext *ClientContext, baseDatatype iface.BaseDatatype) *DatatypeContext {
	logger := log.NewWithTags(constants.TagSdkDatatype, clientContext.Client.GetSummary()+"|"+baseDatatype.GetSummary())
	return &DatatypeContext{
		ClientContext: &ClientContext{
			OrdaContext: &ordaContext{
				Context: clientContext.Ctx(),
				logger:  logger,
			},
			Client: clientContext.Client,
		},
		Data: baseDatatype,
	}
}

// L returns OrdaLog
func (its *ordaContext) L() *log.OrdaLog {
	if its.logger == nil {
		return log.Logger
	}
	return its.logger
}

func (its *ordaContext) Ctx() context.Context {
	return its.Context
}

func (its *ordaContext) SetLogger(l *log.OrdaLog) {
	its.logger = l
}
