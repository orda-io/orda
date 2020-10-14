package context

import (
	"bytes"
	"context"
	"github.com/knowhunger/ortoo/pkg/log"
	"github.com/knowhunger/ortoo/pkg/utils"
)

const maxSubLevelLength = 20

// OrtooContext is a context used in Ortoo
type OrtooContext interface {
	context.Context
	L() *log.OrtooLog
}

type ortooContext struct {
	context.Context
	Logger *log.OrtooLog
}

// New creates a new OrtooContext
func New(ctx context.Context, lv1 mainLevel, lv2, tag string) OrtooContext {
	logger := log.NewWithTag(string(lv1), lv2, tag)
	return &ortooContext{
		Context: ctx,
		Logger:  logger,
	}
}

// NewWithTag creates a new OrtooContext with tag
func NewWithTag(ctx context.Context, lv mainLevel, subLevels ...string) OrtooContext {
	subLevel := ""
	if len(subLevels) >= 1 {
		subLevel = utils.TrimLong(subLevels[0], maxSubLevelLength)
	}
	b := &bytes.Buffer{}
	if len(subLevels) >= 2 {
		b.WriteString(subLevels[1])
	}
	if len(subLevels) >= 3 {
		b.WriteString("|")
		b.WriteString(subLevels[2])
	}
	return New(ctx, lv, subLevel, b.String())
}

// L returns OrtooLog
func (its *ortooContext) L() *log.OrtooLog {
	return its.Logger
}

type mainLevel string

const (
	CLIENT   mainLevel = "CLIE"
	DATATYPE mainLevel = "DATA"
	SERVER   mainLevel = "SERV"
	TEST     mainLevel = "TEST"
)
