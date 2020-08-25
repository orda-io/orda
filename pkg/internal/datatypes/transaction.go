package datatypes

import (
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/iface"
	"github.com/knowhunger/ortoo/pkg/log"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/pkg/operations"
	"sync"
)

// NotUserTransactionTag ...
const NotUserTransactionTag = "NotUserTransactionTag!@#$%ORTOO"

// TransactionDatatype is the datatype responsible for the transaction.
type TransactionDatatype struct {
	*WiredDatatype
	mutex            *sync.RWMutex
	isLocked         bool
	success          bool
	rollbackSnapshot iface.Snapshot
	rollbackOps      []iface.Operation
	rollbackOpID     *model.OperationID
	currentTrxCtx    *TransactionContext
}

// TransactionContext is a context used for transactions
type TransactionContext struct {
	tag          string
	opBuffer     []iface.Operation
	rollbackOpID *model.OperationID
}

func (t *TransactionContext) appendOperation(op iface.Operation) {
	t.opBuffer = append(t.opBuffer, op)
}

// newTransactionDatatype creates a new TransactionDatatype
func newTransactionDatatype(w *WiredDatatype, snapshot iface.Snapshot) *TransactionDatatype {
	return &TransactionDatatype{
		WiredDatatype:    w,
		mutex:            new(sync.RWMutex),
		isLocked:         false,
		success:          true,
		currentTrxCtx:    nil,
		rollbackSnapshot: snapshot.CloneSnapshot(),
		rollbackOps:      nil,
		rollbackOpID:     w.opID.Clone(),
	}
}

// GetWired returns WiredDatatype
func (its *TransactionDatatype) GetWired() *WiredDatatype {
	return its.WiredDatatype
}

// ExecuteOperationWithTransaction is a method to execute an operation with a transaction.
// an operation can be either local or remote
func (its *TransactionDatatype) ExecuteOperationWithTransaction(
	ctx *TransactionContext,
	op iface.Operation,
	isLocal bool,
) (interface{}, errors.OrtooError) {
	transactionCtx := its.BeginTransaction(NotUserTransactionTag, ctx, false)
	defer func() {
		if err := its.EndTransaction(transactionCtx, false, isLocal); err != nil {
			_ = log.OrtooError(err)
		}
	}()

	if isLocal {
		ret, err := its.executeLocalBase(op)
		if err != nil {
			return 0, errors.ErrDatatypeTransaction.New(its.Logger, err.Error())
		}
		its.currentTrxCtx.appendOperation(op)
		return ret, nil
	}
	its.executeRemoteBase(op)
	return nil, nil
}

// make a transaction and lock
func (its *TransactionDatatype) setTransactionContextAndLock(tag string) *TransactionContext {
	if tag != NotUserTransactionTag {
		its.Logger.Infof("Begin the transaction: `%s`", tag)
	}
	its.mutex.Lock()
	its.isLocked = true
	transactionCtx := &TransactionContext{
		tag:          tag,
		opBuffer:     nil,
		rollbackOpID: its.opID.Clone(),
	}
	return transactionCtx
}

// BeginTransaction is called before a transaction is executed.
// This sets TransactionDatatype.currentTrxCtx, lock, and generates a transaction operation
// This is called in either DoTransaction() or ExecuteOperationWithTransaction().
// Note that TransactionDatatype.currentTrxCtx is currently working transaction context.
func (its *TransactionDatatype) BeginTransaction(
	tag string,
	tnxCtx *TransactionContext,
	withOp bool,
) *TransactionContext {
	if its.isLocked && its.currentTrxCtx == tnxCtx {
		return nil // called after DoTransaction() succeeds.
	}
	its.currentTrxCtx = its.setTransactionContextAndLock(tag)
	if withOp {
		op := operations.NewTransactionOperation(tag)
		its.SetNextOpID(op)
		its.currentTrxCtx.appendOperation(op)
	}
	return its.currentTrxCtx
}

// Rollback is called to rollback a transaction
func (its *TransactionDatatype) Rollback() errors.OrtooError {
	its.Logger.Infof("Begin the rollback: '%s'", its.currentTrxCtx.tag)
	snapshotDatatype, _ := its.datatype.(iface.SnapshotDatatype)
	redoOpID := its.opID
	redoSnapshot := snapshotDatatype.GetSnapshot().CloneSnapshot()
	its.SetOpID(its.currentTrxCtx.rollbackOpID)
	snapshotDatatype.SetSnapshot(its.rollbackSnapshot)
	for _, op := range its.rollbackOps {
		err := its.Replay(op)
		if err != nil {
			its.SetOpID(redoOpID)
			snapshotDatatype.SetSnapshot(redoSnapshot)
			return errors.ErrDatatypeTransactionRollback.New(its.Logger, err.Error())
		}
	}
	its.rollbackOpID = its.opID.Clone()
	its.rollbackSnapshot = snapshotDatatype.GetSnapshot().CloneSnapshot()
	its.rollbackOps = nil
	its.Logger.Infof("End the rollback: '%s'", its.currentTrxCtx.tag)
	return nil
}

// SetTransactionFail is called when a transaction fails
func (its *TransactionDatatype) SetTransactionFail() {
	its.success = false
}

// EndTransaction is called when a transaction ends
func (its *TransactionDatatype) EndTransaction(trxCtx *TransactionContext, withOp, isLocal bool) errors.OrtooError {
	if trxCtx == its.currentTrxCtx {
		defer its.unlock()
		if its.success {
			if withOp {
				beginOp, ok := its.currentTrxCtx.opBuffer[0].(*operations.TransactionOperation)
				if !ok {
					return errors.ErrDatatypeTransaction.New(its.Logger, "no transaction operation")
				}
				beginOp.SetNumOfOps(len(its.currentTrxCtx.opBuffer))
			}
			its.rollbackOps = append(its.rollbackOps, its.currentTrxCtx.opBuffer...)
			if isLocal {
				its.deliverTransaction(its.currentTrxCtx.opBuffer)
			}
			if its.currentTrxCtx.tag != NotUserTransactionTag {
				its.Logger.Infof("End the transaction: `%s`", its.currentTrxCtx.tag)
			}
		} else {
			if err := its.Rollback(); err != nil {
				panic(err)
			}

		}
	}
	return nil
}

func (its *TransactionDatatype) unlock() {
	if its.isLocked {
		its.currentTrxCtx = nil
		its.success = true
		its.mutex.Unlock()
		its.isLocked = false
	}
}
