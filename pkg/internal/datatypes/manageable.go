package datatypes

import (
	"encoding/json"
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/iface"
	"github.com/knowhunger/ortoo/pkg/log"
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
) {
	wiredDatatype := newWiredDatatype(base, w)
	transactionDatatype := newTransactionDatatype(wiredDatatype, snapshot)

	its.TransactionDatatype = transactionDatatype
	its.TransactionCtx = nil
	its.SetDatatype(datatype)
}

// GetMeta returns the binary of metadata of the datatype.
func (its *ManageableDatatype) GetMeta() ([]byte, errors.OrtooError) {
	meta := model.DatatypeMeta{
		Key:    its.Key,
		DUID:   its.id,
		OpID:   its.opID,
		TypeOf: its.TypeOf,
		State:  its.state,
	}
	metab, err := json.Marshal(&meta)
	if err != nil {
		return nil, errors.ErrDatatypeMarshal.New(its.Logger, meta)
	}
	return metab, nil
}

// SetMeta sets the metadata with binary metadata.
func (its *ManageableDatatype) SetMeta(meta []byte) errors.OrtooError {
	m := model.DatatypeMeta{}
	if err := json.Unmarshal(meta, &m); err != nil {
		return errors.ErrDatatypeUnmarshal.New(its.Logger, string(meta))
	}
	its.Key = m.Key
	its.id = m.DUID
	its.opID = m.OpID
	its.TypeOf = m.TypeOf
	its.state = m.State
	return nil
}

// DoTransaction enables datatypes to perform a transaction.
func (its *ManageableDatatype) DoTransaction(
	tag string,
	userFunc func(txnCtx *TransactionContext) error,
) errors.OrtooError {
	txnCtx := its.BeginTransaction(tag, its.TransactionCtx, true)
	defer func() {
		if err := its.EndTransaction(txnCtx, true, true); err != nil {
			_ = log.OrtooError(err)
		}
	}()
	if err := userFunc(txnCtx); err != nil {
		its.SetTransactionFail()
		return errors.ErrDatatypeTransaction.New(its.Logger, err.Error())
	}
	return nil
}

// SubscribeOrCreate enables a datatype to subscribe and create itself.
func (its *ManageableDatatype) SubscribeOrCreate(state model.StateOfDatatype) errors.OrtooError {
	if state == model.StateOfDatatype_DUE_TO_SUBSCRIBE {
		its.state = state
		return nil
	}
	subscribeOp, err := operations.NewSnapshotOperation(its.TypeOf, state, its.datatype.GetSnapshot())
	if err != nil {
		return errors.ErrDatatypeSubscribe.New(its.Logger, err.Error())
	}
	_, err = its.ExecuteOperationWithTransaction(its.TransactionCtx, subscribeOp, true)
	if err != nil {
		return errors.ErrDatatypeSubscribe.New(its.Logger, err.Error())
	}
	return nil
}

// ExecuteTransactionRemote is a method to execute a transaction of remote operations
func (its ManageableDatatype) ExecuteRemoteTransaction(
	transaction []*model.Operation,
	obtainList bool,
) ([]interface{}, errors.OrtooError) {
	var trxCtx *TransactionContext
	if len(transaction) > 1 {
		trxOp, ok := operations.ModelToOperation(transaction[0]).(*operations.TransactionOperation)
		if !ok {
			return nil, errors.ErrDatatypeTransaction.New(its.Logger, "no transaction operation")
		}
		if int(trxOp.GetNumOfOps()) != len(transaction) {
			return nil, errors.ErrDatatypeTransaction.New(its.Logger, "not matched number of operations")
		}
		trxCtx = its.BeginTransaction(trxOp.C.Tag, its.TransactionCtx, false)
		defer func() {
			if err := its.EndTransaction(trxCtx, false, false); err != nil {
				_ = log.OrtooError(err)
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
		_, err := its.ExecuteOperationWithTransaction(trxCtx, op, false)
		if err != nil {
			return nil, errors.ErrDatatypeTransaction.New(its.Logger, err.Error())
		}
	}
	return opList, nil
}
