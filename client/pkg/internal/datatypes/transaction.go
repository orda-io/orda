package datatypes

import (
	"encoding/json"
	errors2 "github.com/orda-io/orda/client/pkg/errors"
	"github.com/orda-io/orda/client/pkg/iface"
	"github.com/orda-io/orda/client/pkg/model"
	operations2 "github.com/orda-io/orda/client/pkg/operations"
	"sync"
)

// NotUserTransactionTag ...
const NotUserTransactionTag = "NotUserTransactionTag!@#$%Orda"

// TransactionContext is a context used for transactions
type TransactionContext struct {
	tag      string
	opBuffer []iface.Operation
}

func (its *TransactionContext) appendOperation(op iface.Operation) {
	its.opBuffer = append(its.opBuffer, op)
}

// TransactionDatatype is the datatype responsible for the transaction.
type TransactionDatatype struct {
	*BaseDatatype
	mutex            *sync.RWMutex
	isLocked         bool
	success          bool
	rollbackSnapshot []byte
	rollbackMeta     []byte
	rollbackOps      []iface.Operation
	txCtx            *TransactionContext
}

// NewTransactionDatatype creates a new TransactionDatatype
func NewTransactionDatatype(b *BaseDatatype) *TransactionDatatype {

	return &TransactionDatatype{
		BaseDatatype: b,
		mutex:        new(sync.RWMutex),
		isLocked:     false,
		success:      true,
		txCtx:        nil,
		rollbackOps:  nil,
	}
}

func (its *TransactionDatatype) ResetTransaction() errors2.OrdaError {
	snap, err := json.Marshal(its.GetSnapshot())
	if err != nil {
		return errors2.DatatypeMarshal.New(its.L(), err.Error())
	}
	meta, err := its.GetMeta()
	if err != nil {
		return errors2.DatatypeMarshal.New(its.L(), err.Error())
	}
	its.rollbackSnapshot = snap
	its.rollbackMeta = meta
	its.rollbackOps = nil
	return nil
}

// SentenceInTx is a method to execute an operation with a transaction.
// an operation can be either local or remote
func (its *TransactionDatatype) SentenceInTx(
	ctx *TransactionContext,
	op iface.Operation,
	isLocal bool,
) (interface{}, errors2.OrdaError) {
	transactionCtx := its.BeginTransaction(NotUserTransactionTag, ctx, false)
	defer func() {
		if err := its.EndTransaction(transactionCtx, false, isLocal); err != nil {

		}
	}()
	its.ctx.L().Infof("sentence: %+v", op)
	if isLocal {
		ret, err := its.executeLocalBase(op)
		if err != nil {
			return ret, err
		}
		its.txCtx.appendOperation(op)
		return ret, nil
	}
	its.executeRemoteBase(op)
	its.txCtx.appendOperation(op)
	return nil, nil
}

// make a transaction and lock
func (its *TransactionDatatype) setTransactionContextAndLock(tag string) *TransactionContext {
	if tag != NotUserTransactionTag {
		its.L().Infof("Begin the transaction: '%s'", tag)
	}
	its.mutex.Lock()
	its.isLocked = true
	return &TransactionContext{
		tag:      tag,
		opBuffer: nil,
	}
}

// BeginTransaction is called before a transaction is executed.
// This sets TransactionDatatype.txCtx, lock, and generates a transaction operation
// This is called in either DoTransaction() or SentenceInTx().
// Note that TransactionDatatype.txCtx is currently working transaction context.
func (its *TransactionDatatype) BeginTransaction(
	tag string,
	txCtx *TransactionContext,
	newTxnOp bool,
) *TransactionContext {
	if its.isLocked && its.txCtx == txCtx {
		return nil // called after DoTransaction() succeeds.
	}
	its.txCtx = its.setTransactionContextAndLock(tag)
	if newTxnOp {
		op := operations2.NewTransactionOperation(tag)
		its.SetNextOpID(op)
		its.txCtx.appendOperation(op)
	}
	return its.txCtx
}

// Rollback is called to rollback a transaction
func (its *TransactionDatatype) Rollback() errors2.OrdaError {
	its.L().Infof("Begin the rollback: '%s'", its.txCtx.tag)
	if err := its.SetMetaAndSnapshot(its.rollbackMeta, its.rollbackSnapshot); err != nil {
		return errors2.DatatypeTransaction.New(its.L(), "rollback failed")
	}
	for _, op := range its.rollbackOps {
		if err := its.Replay(op); err != nil {
			return errors2.DatatypeTransaction.New(its.L(), "rollback failed")
		}
	}
	var err errors2.OrdaError
	if its.rollbackMeta, its.rollbackSnapshot, err = its.GetMetaAndSnapshot(); err != nil {
		return errors2.DatatypeTransaction.New(its.L(), "rollback failed")
	}
	its.rollbackOps = nil
	its.L().Infof("End the rollback: '%s'", its.txCtx.tag)
	return nil
}

// SetTransactionFail is called when a transaction fails
func (its *TransactionDatatype) SetTransactionFail() {
	its.success = false
}

// EndTransaction is called when a transaction ends
func (its *TransactionDatatype) EndTransaction(txCtx *TransactionContext, withOp, isLocal bool) errors2.OrdaError {
	if txCtx == its.txCtx {
		defer its.unlock()
		if its.success {
			if withOp {
				beginOp, ok := its.txCtx.opBuffer[0].(*operations2.TransactionOperation)
				if !ok {
					return errors2.DatatypeTransaction.New(its.L(), "no transaction operation")
				}
				beginOp.SetNumOfOps(len(its.txCtx.opBuffer))
			}
			its.rollbackOps = append(its.rollbackOps, its.txCtx.opBuffer...)
			if isLocal {
				its.DeliverTransaction(its.txCtx.opBuffer)
			}
			if its.txCtx.tag != NotUserTransactionTag {
				its.L().Infof("End the transaction: `%s`", its.txCtx.tag)
			}
		} else if err := its.Rollback(); err != nil {
			panic(err)
		}
	}
	return nil
}

func (its *TransactionDatatype) unlock() {
	if its.isLocked {
		its.txCtx = nil
		its.success = true
		its.mutex.Unlock()
		its.isLocked = false
	}
}

// DoTransaction enables datatypes to perform a transaction.
func (its *TransactionDatatype) DoTransaction(
	tag string,
	currentTxCtx *TransactionContext,
	funcWithCloneDatatype func(txCtx *TransactionContext) error,
) errors2.OrdaError {
	txCtx := its.BeginTransaction(tag, currentTxCtx, true)
	defer func() {
		if err := its.EndTransaction(txCtx, true, true); err != nil {
			// do nothing
		}
	}()
	if err := funcWithCloneDatatype(txCtx); err != nil {
		its.SetTransactionFail()
		return errors2.DatatypeTransaction.New(its.L(), err.Error())
	}
	return nil
}

// ExecuteRemoteTransactionWithCtx is a method to execute a transaction of remote operations
func (its *TransactionDatatype) ExecuteRemoteTransactionWithCtx(
	transaction []*model.Operation,
	currentTxCtx *TransactionContext,
	obtainList bool,
) ([]interface{}, errors2.OrdaError) {
	var txCtx *TransactionContext
	if len(transaction) > 1 {
		txOp, ok := operations2.ModelToOperation(transaction[0]).(*operations2.TransactionOperation)
		if !ok {
			return nil, errors2.DatatypeTransaction.New(its.L(), "no transaction operation")
		}
		if int(txOp.GetNumOfOps()) != len(transaction) {
			return nil, errors2.DatatypeTransaction.New(its.L(), "not matched number of operations")
		}
		txCtx = its.BeginTransaction(txOp.GetBody().Tag, currentTxCtx, false)
		defer func() {
			if err := its.EndTransaction(txCtx, false, false); err != nil {
				// _ = log.OrdaError(err)
			}
		}()
		transaction = transaction[1:]
	}
	var opList []interface{}
	for _, modelOp := range transaction {
		op := operations2.ModelToOperation(modelOp)
		if obtainList {
			opList = append(opList, op.ToJSON())
		}
		if _, err := its.SentenceInTx(txCtx, op, false); err != nil {
			return nil, errors2.DatatypeTransaction.New(its.L(), err.Error())
		}
	}
	return opList, nil
}
