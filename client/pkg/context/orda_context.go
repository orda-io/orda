package context

import (
	"context"
	"github.com/orda-io/orda/client/pkg/iface"
	"github.com/orda-io/orda/client/pkg/log"
	"strconv"
)

type ordaContext struct {
	context.Context
	logger *log.OrdaLog
}

// UpdateAllTags updates all tags; if the specified tags are empty, previous tags are kept.
func (its *ordaContext) UpdateAllTags(emoji, collection, colNum, client, cuid, datatype, duid string) {
	its.logger.SetTags(emoji, collection, colNum, client, cuid, datatype, duid)
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

// UpdateCollectionTags updates the collection tag
func (its *ordaContext) UpdateCollectionTags(collection string, colNum int32) iface.OrdaContext {
	cn := ""
	if colNum > 0 {
		cn = strconv.Itoa(int(colNum))
	}
	its.UpdateAllTags("", collection, cn, "", "", "", "")
	return its
}

// UpdateClientTags updates the client tag
func (its *ordaContext) UpdateClientTags(client, cuid string) iface.OrdaContext {
	its.UpdateAllTags("", "", "", client, cuid, "", "")
	return its
}

// UpdateDatatypeTags updates the datatype tag
func (its *ordaContext) UpdateDatatypeTags(datatype, duid string) iface.OrdaContext {
	its.UpdateAllTags("", "", "", "", "", datatype, duid)
	return its
}

func (its *ordaContext) CloneWithNewEmoji(emoji string) iface.OrdaContext {
	newLogger := its.logger.Clone()
	newLogger.SetTags(emoji, "", "", "", "", "", "")
	return &ordaContext{
		Context: context.TODO(),
		logger:  newLogger,
	}
}

func NewOrdaContext(ctx context.Context, emoji string) iface.OrdaContext {
	return NewOrdaContextWithAllTags(ctx, emoji, "", "", "", "", "", "")
}

// NewOrdaContextWithAllTags creates a new OrdaContext
func NewOrdaContextWithAllTags(ctx context.Context, emoji, collection, colNum, client, cuid, datatype, duid string) iface.OrdaContext {
	logger := log.NewWithTags(emoji, collection, colNum, client, cuid, datatype, duid)
	return &ordaContext{
		Context: ctx,
		logger:  logger,
	}
}
