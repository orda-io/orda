package context

import (
	"context"
	"github.com/knowhunger/ortoo/pkg/log"
)

// OrtooContext is a context used in Ortoo
type OrtooContext interface {
	context.Context
	L() *log.OrtooLog
	SetNewLogger(lv mainTag, subLevel string)
	SetLogger(l *log.OrtooLog)
}

type ortooContext struct {
	context.Context
	Logger *log.OrtooLog
}

func New(ctx context.Context) OrtooContext {
	return &ortooContext{
		Context: ctx,
	}
}

// New creates a new OrtooContext
func NewWithTags(ctx context.Context, tag1 mainTag, tag2 string) OrtooContext {
	logger := log.NewWithTags(string(tag1), tag2)
	return &ortooContext{
		Context: ctx,
		Logger:  logger,
	}
}

// L returns OrtooLog
func (its *ortooContext) L() *log.OrtooLog {
	if its.Logger == nil {
		return log.Logger
	}
	return its.Logger
}

func (its *ortooContext) SetNewLogger(lv1 mainTag, lv2 string) {
	its.Logger = log.NewWithTags(string(lv1), lv2)
}

func (its *ortooContext) SetLogger(l *log.OrtooLog) {
	its.Logger = l
}

type mainTag string

const (
	CLIENT   mainTag = "CLIE"
	DATATYPE mainTag = "DATA"
	SERVER   mainTag = "SERV"
	TEST     mainTag = "TEST"
)
