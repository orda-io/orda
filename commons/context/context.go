package context

import (
	"context"
	"fmt"
	"github.com/knowhunger/ortoo/commons/log"
)

type OrtooContext struct {
	context.Context
	Logger *log.OrtooLog
}

func NewOrtooContext() *OrtooContext {
	context := context.TODO()
	logger := log.NewOrtooLogWithTag(fmt.Sprintf("%p", context))
	return &OrtooContext{
		Context: context,
		Logger:  logger,
	}
}
