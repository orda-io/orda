package context

import (
	"context"
	"fmt"
	"github.com/knowhunger/ortoo/commons/log"
)

// OrtooContext is a context used in Ortoo
type OrtooContext struct {
	context.Context
	Logger *log.OrtooLog
}

// NewOrtooContext creates a new Ortoo context.
func NewOrtooContext() *OrtooContext {
	context := context.TODO()
	logger := log.NewOrtooLogWithTag(fmt.Sprintf("%p", context))
	return &OrtooContext{
		Context: context,
		Logger:  logger,
	}
}
