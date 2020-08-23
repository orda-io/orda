package ortoo

import (
	"github.com/knowhunger/ortoo/ortoo/errors"
	"github.com/knowhunger/ortoo/ortoo/model"
)

// Handlers defines a set of handlers which can handles the events related to Datatype
type Handlers struct {
	stateChangeHandler     func(dt Datatype, old model.StateOfDatatype, new model.StateOfDatatype)
	remoteOperationHandler func(dt Datatype, opList []interface{})
	errorHandler           func(dt Datatype, errs ...errors.OrtooError)
}

// NewHandlers creates a set of handlers for a datatype.
func NewHandlers(
	stateChangeHandler func(dt Datatype, old model.StateOfDatatype, new model.StateOfDatatype),
	remoteOperationHandler func(dt Datatype, opList []interface{}),
	errorHandler func(dt Datatype, errs ...errors.OrtooError)) *Handlers {
	return &Handlers{
		stateChangeHandler:     stateChangeHandler,
		remoteOperationHandler: remoteOperationHandler,
		errorHandler:           errorHandler,
	}
}

// SetHandlers sets the handlers if a given handler is not nil.
func (its *Handlers) SetHandlers(
	stateChangeHandler func(dt Datatype, old model.StateOfDatatype, new model.StateOfDatatype),
	remoteOperationHandler func(dt Datatype, opList []interface{}),
	errorHandler func(dt Datatype, errs ...errors.OrtooError)) {
	if stateChangeHandler != nil {
		its.stateChangeHandler = stateChangeHandler
	}
	if remoteOperationHandler != nil {
		its.remoteOperationHandler = remoteOperationHandler
	}
	if errorHandler != nil {
		its.errorHandler = errorHandler
	}
}
