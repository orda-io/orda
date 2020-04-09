package ortoo

import (
	"github.com/knowhunger/ortoo/ortoo/iface"
	"github.com/knowhunger/ortoo/ortoo/internal/datatypes"
	"github.com/knowhunger/ortoo/ortoo/model"
)

// Datatype is an Ortoo Datatype which provides common interfaces.
type Datatype interface {
	iface.PublicBaseDatatype
	iface.PublicSnapshotDatatype
}

type datatype struct {
	*datatypes.ManageableDatatype
	handlers *Handlers
}

func (its *datatype) HandleStateChange(old, new model.StateOfDatatype) {
	if its.handlers != nil && its.handlers.stateChangeHandler != nil {
		go its.handlers.stateChangeHandler(its.GetDatatype().(Datatype), old, new)
	}
}

func (its *datatype) HandleErrors(errs ...error) {
	if its.handlers != nil && its.handlers.errorHandler != nil {
		go its.handlers.errorHandler(its.ManageableDatatype.GetDatatype().(Datatype), errs...)
	}
}

func (its *datatype) HandleRemoteOperations(operations []interface{}) {
	if its.handlers != nil && its.handlers.remoteOperationHandler != nil {
		go its.handlers.remoteOperationHandler(its.ManageableDatatype.GetDatatype().(Datatype), operations)
	}
}
