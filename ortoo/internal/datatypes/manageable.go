package datatypes

import (
	"encoding/json"
	"github.com/knowhunger/ortoo/ortoo/errors"
	"github.com/knowhunger/ortoo/ortoo/iface"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/ortoo/model"
	"github.com/knowhunger/ortoo/ortoo/operations"
	"github.com/knowhunger/ortoo/ortoo/types"
)

// ManageableDatatype implements the datatype features finally used.
type ManageableDatatype struct {
	*TransactionDatatype
	TransactionCtx *TransactionContext
}

// Initialize is a method for initialization
func (its *ManageableDatatype) Initialize(
	key string,
	typeOf model.TypeOfDatatype,
	cuid types.CUID,
	w iface.Wire,
	snapshot iface.Snapshot,
	datatype iface.Datatype) {

	baseDatatype := newBaseDatatype(key, typeOf, cuid)
	wiredDatatype := newWiredDatatype(baseDatatype, w)
	transactionDatatype := newTransactionDatatype(wiredDatatype, snapshot)

	its.TransactionDatatype = transactionDatatype
	its.TransactionCtx = nil
	its.SetDatatype(datatype)
}

// GetMeta returns the binary of metadata of the datatype.
func (its *ManageableDatatype) GetMeta() ([]byte, error) {
	meta := model.DatatypeMeta{
		Key:    its.Key,
		DUID:   its.id,
		OpID:   its.opID,
		TypeOf: its.TypeOf,
		State:  its.state,
	}
	metab, err := json.Marshal(&meta)
	if err != nil {
		return nil, log.OrtooError(err)
	}
	return metab, nil
}

// SetMeta sets the metadata with binary metadata.
func (its *ManageableDatatype) SetMeta(meta []byte) error {
	m := model.DatatypeMeta{}
	if err := json.Unmarshal(meta, &m); err != nil {
		return log.OrtooError(err)
	}
	its.Key = m.Key
	its.id = m.DUID
	its.opID = m.OpID
	its.TypeOf = m.TypeOf
	its.state = m.State
	return nil
}

// DoTransaction enables datatypes to perform a transaction.
func (its *ManageableDatatype) DoTransaction(tag string, fn func(txnCtx *TransactionContext) error) error {
	txnCtx, err := its.BeginTransaction(tag, its.TransactionCtx, true)
	if err != nil {
		return err
	}
	defer func() {
		if err := its.EndTransaction(txnCtx, true, true); err != nil {
			_ = log.OrtooError(err)
		}
	}()
	if err := fn(txnCtx); err != nil {
		its.SetTransactionFail()
		return errors.New(errors.ErrDatatypeTransaction, err.Error())
	}
	return nil
}

// SubscribeOrCreate enables a datatype to subscribe and create itself.
func (its *ManageableDatatype) SubscribeOrCreate(state model.StateOfDatatype) error {
	if state == model.StateOfDatatype_DUE_TO_SUBSCRIBE {
		its.state = state
		return nil
	}
	subscribeOp, err := operations.NewSnapshotOperation(its.TypeOf, state, its.datatype.GetSnapshot())
	if err != nil {
		return log.OrtooErrorf(err, "fail to subscribe")
	}
	_, err = its.ExecuteOperationWithTransaction(its.TransactionCtx, subscribeOp, true)
	if err != nil {
		return log.OrtooErrorf(err, "fail to execute SubscribeOperation")
	}
	return nil
}

// ExecuteTransactionRemote is a method to execute a transaction of remote operations
func (its ManageableDatatype) ExecuteTransactionRemote(transaction []*model.Operation, obtainList bool) ([]interface{}, error) {
	var trxCtx *TransactionContext
	var err error
	if len(transaction) > 1 {
		trxOp, ok := operations.ModelToOperation(transaction[0]).(*operations.TransactionOperation)
		if !ok {
			return nil, errors.New(errors.ErrDatatypeTransaction, "no transaction operation")
		}
		if int(trxOp.GetNumOfOps()) != len(transaction) {
			return nil, errors.New(errors.ErrDatatypeTransaction, "not matched number of operations")
		}
		trxCtx, err = its.BeginTransaction(trxOp.C.Tag, its.TransactionCtx, false)
		if err != nil {
			return nil, log.OrtooError(err)
		}
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
			return nil, errors.New(errors.ErrDatatypeTransaction, err.Error())
		}
	}
	return opList, nil
}
