package iface

import (
	"context"
	"github.com/orda-io/orda/client/pkg/log"
)

// OrdaContext is used to pass over the context of clients or datatypes
type OrdaContext interface {
	context.Context
	L() *log.OrdaLog
	Ctx() context.Context
	SetLogger(l *log.OrdaLog)
	UpdateAllTags(emoji, collection, colnum, client, cuid, datatype, duid string)
	UpdateCollectionTags(collection string, colNum int32) OrdaContext
	UpdateClientTags(client, cuid string) OrdaContext
	UpdateDatatypeTags(datatype, duid string) OrdaContext
	CloneWithNewEmoji(emoji string) OrdaContext
}
