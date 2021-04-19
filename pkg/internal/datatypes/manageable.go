package datatypes

import (
	"encoding/json"
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/iface"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/pkg/operations"
)

// ManageableDatatype implements the datatype features finally used.
type ManageableDatatype struct {
	*TransactionDatatype
	TransactionCtx *TransactionContext
}

// Initialize is a method for initialization
func (its *ManageableDatatype) Initialize(
	base *BaseDatatype,
	w iface.Wire,
	snapshot iface.Snapshot,
	datatype iface.Datatype,
) errors.OrtooError {
	wiredDatatype := newWiredDatatype(base, w)
	transactionDatatype := newTransactionDatatype(wiredDatatype, snapshot)
	its.TransactionDatatype = transactionDatatype
	its.TransactionCtx = nil
	its.SetDatatype(datatype)
	if err := its.ResetRollBackContext(); err != nil {
		return err
	}
	return nil
}

// DoTransaction enables datatypes to perform a transaction.
func (its *ManageableDatatype) DoTransaction(
	tag string,
	funcWithCloneDatatype func(txCtx *TransactionContext) error,
) errors.OrtooError {
	txCtx := its.BeginTransaction(tag, its.TransactionCtx, true)
	defer func() {
		if err := its.EndTransaction(txCtx, true, true); err != nil {
			// do nothing
		}
	}()
	if err := funcWithCloneDatatype(txCtx); err != nil {
		its.SetTransactionFail()
		return errors.DatatypeTransaction.New(its.L(), err.Error())
	}
	return nil
}

// SubscribeOrCreate enables a datatype to subscribe and create itself.
func (its *ManageableDatatype) SubscribeOrCreate(state model.StateOfDatatype) errors.OrtooError {
	if state == model.StateOfDatatype_DUE_TO_SUBSCRIBE {
		its.state = state
		return nil
	}
	snap, err := json.Marshal(its.datatype.GetSnapshot())
	if err != nil {
		return errors.DatatypeSubscribe.New(its.L(), err.Error())
	}
	subscribeOp := operations.NewSnapshotOperation(its.TypeOf, state, string(snap))
	_, err = its.SentenceInTransaction(its.TransactionCtx, subscribeOp, true)
	if err != nil {
		return errors.DatatypeSubscribe.New(its.L(), err.Error())
	}
	return nil
}

// ExecuteRemoteTransaction is a method to execute a transaction of remote operations
func (its ManageableDatatype) ExecuteRemoteTransaction(
	transaction []*model.Operation,
	obtainList bool,
) ([]interface{}, errors.OrtooError) {
	var txCtx *TransactionContext
	if len(transaction) > 1 {
		txOp, ok := operations.ModelToOperation(transaction[0]).(*operations.TransactionOperation)
		if !ok {
			return nil, errors.DatatypeTransaction.New(its.L(), "no transaction operation")
		}
		if int(txOp.GetNumOfOps()) != len(transaction) {
			return nil, errors.DatatypeTransaction.New(its.L(), "not matched number of operations")
		}
		txCtx = its.BeginTransaction(txOp.C.Tag, its.TransactionCtx, false)
		defer func() {
			if err := its.EndTransaction(txCtx, false, false); err != nil {
				// _ = log.OrtooError(err)
			}
		}()
		transaction = transaction[1:]
	}
	var opList []interface{}
	for _, modelOp := range transaction {
		op := operations.ModelToOperation(modelOp)
		if obtainList {
			opList = append(opList, op.GetAsJSON())
		}
		_, err := its.SentenceInTransaction(txCtx, op, false)
		if err != nil {
			return nil, errors.DatatypeTransaction.New(its.L(), err.Error())
		}
	}
	return opList, nil
}
