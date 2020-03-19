package ortoo

import (
	"github.com/knowhunger/ortoo/ortoo/internal/datatypes"
	"github.com/knowhunger/ortoo/ortoo/model"
)

type Datatype interface {
	datatypes.PublicWiredDatatypeInterface
}

type datatype struct {
	*datatypes.FinalDatatype
	handlers *Handlers
}

func (its *datatype) HandleStateChange(old, new model.StateOfDatatype) {
	if its.handlers != nil && its.handlers.stateChangeHandler != nil {
		go its.handlers.stateChangeHandler(its.FinalDatatype.GetDatatype().(Datatype), old, new)
	}
}

func (its *datatype) HandleErrors(errs ...error) {

	if its.handlers != nil && its.handlers.errorHandler != nil {
		go its.handlers.errorHandler(its.FinalDatatype.GetDatatype().(Datatype), errs...)
	}
}

func (its *datatype) HandleRemoteOperations(operations []interface{}) {
	if its.handlers != nil && its.handlers.remoteOperationHandler != nil {
		go its.handlers.remoteOperationHandler(its.FinalDatatype.GetDatatype().(Datatype), operations)
	}
}
