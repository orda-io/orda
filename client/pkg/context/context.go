package context

import (
	"context"
	"github.com/orda-io/orda/client/pkg/constants"
	"github.com/orda-io/orda/client/pkg/iface"
	"github.com/orda-io/orda/client/pkg/log"
	"github.com/orda-io/orda/client/pkg/model"
)

type ordaContext struct {
	context.Context
	logger *log.OrdaLog
}

// UpdateTags updates tags
func (its *ordaContext) UpdateTags(tag1, tag2 string) {
	its.logger.SetTags(tag1, tag2)
}

// ClientContext is used to pass over the context of clients
type ClientContext struct {
	iface.OrdaContext
	Client *model.Client
}

// DatatypeContext is used to pass over the context of datatypes
type DatatypeContext struct {
	*ClientContext
	Data iface.BaseDatatype
}

// NewOrdaContext creates a new OrdaContext
func NewOrdaContext(ctx context.Context, tag1 string, tag2 string) iface.OrdaContext {
	logger := log.NewWithTags(tag1, tag2)
	return &ordaContext{
		Context: ctx,
		logger:  logger,
	}
}

// NewClientContext creates a new ClientContext
func NewClientContext(ctx context.Context, client *model.Client) *ClientContext {
	return &ClientContext{
		OrdaContext: NewOrdaContext(ctx, constants.TagSdkClient, client.GetSummary()),
		Client:      client,
	}
}

// NewDatatypeContext creates a new DatatypeContext
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

// Ctx returns the context.Context
func (its *ordaContext) Ctx() context.Context {
	return its.Context
}

// SetLogger sets the logger
func (its *ordaContext) SetLogger(l *log.OrdaLog) {
	its.logger = l
}
