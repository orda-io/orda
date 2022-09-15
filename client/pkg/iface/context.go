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
	UpdateTags(tag1, tag2 string)
}
