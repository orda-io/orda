package context

import (
	"context"
	"github.com/knowhunger/ortoo/pkg/log"
)

// OrtooContext is a context used in Ortoo
type OrtooContext struct {
	context.Context
	Logger *log.OrtooLog
}

// NewOrtooContext creates a new Ortoo context.
func NewOrtooContext(alias string) *OrtooContext {
	context := context.TODO()
	logger := log.NewOrtooLogWithTag(alias)
	return &OrtooContext{
		Context: context,
		Logger:  logger,
	}
}
