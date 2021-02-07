package ortoo

import (
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/iface"
	"github.com/knowhunger/ortoo/pkg/internal/datatypes"
	"github.com/knowhunger/ortoo/pkg/model"
)

type base iface.BaseDatatype // not to export in Snapshot

// Datatype is an Ortoo Datatype which provides common interfaces.
type Datatype interface {
	iface.PublicBaseDatatype
	iface.PublicSnapshotDatatype
}

type datatype struct {
	*datatypes.ManageableDatatype
	handlers *Handlers
}

func (its *datatype) newDatatype(txCtx *datatypes.TransactionContext) *datatype {
	return &datatype{
		ManageableDatatype: &datatypes.ManageableDatatype{
			TransactionDatatype: its.ManageableDatatype.TransactionDatatype,
			TransactionCtx:      txCtx,
		},
		handlers: its.handlers,
	}
}

func (its *datatype) HandleStateChange(old, new model.StateOfDatatype) {
	if its.handlers != nil && its.handlers.stateChangeHandler != nil {
		go its.handlers.stateChangeHandler(its.GetDatatype().(Datatype), old, new)
	}
}

func (its *datatype) HandleErrors(errs ...errors.OrtooError) {
	if its.handlers != nil && its.handlers.errorHandler != nil {
		go its.handlers.errorHandler(its.ManageableDatatype.GetDatatype().(Datatype), errs...)
	}
}

func (its *datatype) HandleRemoteOperations(operations []interface{}) {
	if its.handlers != nil && its.handlers.remoteOperationHandler != nil {
		go its.handlers.remoteOperationHandler(its.ManageableDatatype.GetDatatype().(Datatype), operations)
	}
}
