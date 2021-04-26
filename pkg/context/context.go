package context

import (
	"context"
	"github.com/knowhunger/ortoo/pkg/constants"
	"github.com/knowhunger/ortoo/pkg/iface"
	"github.com/knowhunger/ortoo/pkg/log"
	"github.com/knowhunger/ortoo/pkg/model"
)

// OrtooContext is a context used in Ortoo
type OrtooContext interface {
	context.Context
	L() *log.OrtooLog
	Ctx() context.Context
	SetLogger(l *log.OrtooLog)
	UpdateTags(tag1, tag2 string)
}

type ortooContext struct {
	context.Context
	logger *log.OrtooLog
}

func (its *ortooContext) UpdateTags(tag1, tag2 string) {
	its.logger.SetTags(tag1, tag2)
}

type ClientContext struct {
	OrtooContext
	Client *model.Client
}

type DatatypeContext struct {
	*ClientContext
	Data iface.BaseDatatype
}

func NewOrtooContext(ctx context.Context, tag1 string, tag2 string) OrtooContext {
	logger := log.NewWithTags(tag1, tag2)
	return &ortooContext{
		Context: ctx,
		logger:  logger,
	}
}

func NewClientContext(ctx context.Context, client *model.Client) *ClientContext {
	return &ClientContext{
		OrtooContext: NewOrtooContext(ctx, constants.TagSdkClient, client.GetSummary()),
		Client:       client,
	}
}

func NewDatatypeContext(clientContext *ClientContext, baseDatatype iface.BaseDatatype) *DatatypeContext {
	logger := log.NewWithTags(constants.TagSdkDatatype, clientContext.Client.GetSummary()+"|"+baseDatatype.GetSummary())
	return &DatatypeContext{
		ClientContext: &ClientContext{
			OrtooContext: &ortooContext{
				Context: clientContext.Ctx(),
				logger:  logger,
			},
			Client: clientContext.Client,
		},
		Data: baseDatatype,
	}
}

// L returns OrtooLog
func (its *ortooContext) L() *log.OrtooLog {
	if its.logger == nil {
		return log.Logger
	}
	return its.logger
}

func (its *ortooContext) Ctx() context.Context {
	return its.Context
}

func (its *ortooContext) SetLogger(l *log.OrtooLog) {
	its.logger = l
}
