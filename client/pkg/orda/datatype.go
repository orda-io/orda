package orda

import (
	errors2 "github.com/orda-io/orda/client/pkg/errors"
	iface2 "github.com/orda-io/orda/client/pkg/iface"
	datatypes2 "github.com/orda-io/orda/client/pkg/internal/datatypes"
	model2 "github.com/orda-io/orda/client/pkg/model"
)

// Datatype is an Orda Datatype which provides common interfaces.
type Datatype interface {
	GetType() model2.TypeOfDatatype
	GetState() model2.StateOfDatatype
	GetKey() string // @baseDatatype
	ToJSON() interface{}
}

type datatype struct {
	*datatypes2.WiredDatatype
	TxCtx    *datatypes2.TransactionContext
	handlers *Handlers
}

func newDatatype(
	base *datatypes2.BaseDatatype,
	wire iface2.Wire,
	handlers *Handlers,
) *datatype {
	t := datatypes2.NewTransactionDatatype(base)
	w := datatypes2.NewWiredDatatype(wire, t)
	return &datatype{
		WiredDatatype: w,
		TxCtx:         nil,
		handlers:      handlers,
	}
}

func (its *datatype) init(data iface2.Datatype) errors2.OrdaError {
	its.Datatype = data
	its.ResetWired()
	its.ResetSnapshot()
	return its.ResetTransaction()
}

func (its *datatype) cloneDatatype(txCtx *datatypes2.TransactionContext) *datatype {
	return &datatype{
		WiredDatatype: its.WiredDatatype,
		TxCtx:         txCtx,
		handlers:      its.handlers,
	}
}

func (its *datatype) HandleStateChange(old, new model2.StateOfDatatype) {
	if its.handlers != nil && its.handlers.stateChangeHandler != nil {
		its.handlers.stateChangeHandler(its.Datatype, old, new)
	}
}

func (its *datatype) HandleErrors(errs ...errors2.OrdaError) {
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
func (its *datatype) SubscribeOrCreate(state model2.StateOfDatatype) errors2.OrdaError {
	if state == model2.StateOfDatatype_DUE_TO_SUBSCRIBE {
		its.DeliverTransaction(nil)
		return nil
	}
	snapOp, err := its.CreateSnapshotOperation()
	if err != nil {
		return errors2.DatatypeSubscribe.New(its.L(), err.Error())
	}
	_, err = its.SentenceInTx(its.TxCtx, snapOp, true)
	if err != nil {
		return errors2.DatatypeSubscribe.New(its.L(), err.Error())
	}
	return nil
}

func (its *datatype) ExecuteRemoteTransaction(
	transaction []*model2.Operation,
	obtainList bool,
) ([]interface{}, errors2.OrdaError) {
	return its.ExecuteRemoteTransactionWithCtx(transaction, its.TxCtx, obtainList)
}
