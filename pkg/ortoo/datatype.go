package ortoo

import (
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/iface"
	"github.com/knowhunger/ortoo/pkg/internal/datatypes"
	"github.com/knowhunger/ortoo/pkg/model"
)

// Datatype is an Ortoo Datatype which provides common interfaces.
type Datatype interface {
	GetType() model.TypeOfDatatype
	GetState() model.StateOfDatatype
	GetKey() string // @baseDatatype
	ToJSON() interface{}
}

type datatype struct {
	*datatypes.WiredDatatype
	TxCtx    *datatypes.TransactionContext
	handlers *Handlers
}

func newDatatype(
	base *datatypes.BaseDatatype,
	wire iface.Wire,
	handlers *Handlers,
) *datatype {
	t := datatypes.NewTransactionDatatype(base)
	w := datatypes.NewWiredDatatype(wire, t)
	return &datatype{
		WiredDatatype: w,
		TxCtx:         nil,
		handlers:      handlers,
	}
}

func (its *datatype) init(data iface.Datatype) errors.OrtooError {
	its.Datatype = data
	its.ResetWired()
	its.ResetSnapshot()
	return its.ResetTransaction()
}

func (its *datatype) cloneDatatype(txCtx *datatypes.TransactionContext) *datatype {
	return &datatype{
		WiredDatatype: its.WiredDatatype,
		TxCtx:         txCtx,
		handlers:      its.handlers,
	}
}

func (its *datatype) HandleStateChange(old, new model.StateOfDatatype) {
	if its.handlers != nil && its.handlers.stateChangeHandler != nil {
		its.handlers.stateChangeHandler(its.Datatype, old, new)
	}
}

func (its *datatype) HandleErrors(errs ...errors.OrtooError) {
	if its.handlers != nil && its.handlers.errorHandler != nil {
		its.handlers.errorHandler(its.Datatype, errs...)
	}
}

func (its *datatype) HandleRemoteOperations(operations []interface{}) {
	if its.handlers != nil && its.handlers.remoteOperationHandler != nil {
		its.handlers.remoteOperationHandler(its.Datatype, operations)
	}
}

// SubscribeOrCreate enables a datatype to subscribe and create itself.
func (its *datatype) SubscribeOrCreate(state model.StateOfDatatype) errors.OrtooError {
	if state == model.StateOfDatatype_DUE_TO_SUBSCRIBE {
		its.DeliverTransaction(nil)
		return nil
	}
	snapOp, err := its.CreateSnapshotOperation()
	if err != nil {
		return errors.DatatypeSubscribe.New(its.L(), err.Error())
	}
	_, err = its.SentenceInTx(its.TxCtx, snapOp, true)
	if err != nil {
		return errors.DatatypeSubscribe.New(its.L(), err.Error())
	}
	return nil
}

func (its *datatype) ExecuteRemoteTransaction(
	transaction []*model.Operation,
	obtainList bool,
) ([]interface{}, errors.OrtooError) {
	return its.ExecuteRemoteTransactionWithCtx(transaction, its.TxCtx, obtainList)
}
