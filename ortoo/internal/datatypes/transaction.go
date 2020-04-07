package datatypes

import (
	"github.com/knowhunger/ortoo/ortoo/errors"
	operations "github.com/knowhunger/ortoo/ortoo/operations"

	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/ortoo/model"
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
	rollbackSnapshot model.Snapshot
	rollbackOps      []operations.Operation
	rollbackOpID     *model.OperationID
	currentTrxCtx    *TransactionContext
}

// TransactionContext is a context used for transactions
type TransactionContext struct {
	tag          string
	opBuffer     []operations.Operation
	rollbackOpID *model.OperationID
}

func (t *TransactionContext) appendOperation(op operations.Operation) {
	t.opBuffer = append(t.opBuffer, op)
}

// newTransactionDatatype creates a new TransactionDatatype
func newTransactionDatatype(w *WiredDatatype, snapshot model.Snapshot) *TransactionDatatype {
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
func (t *TransactionDatatype) GetWired() *WiredDatatype {
	return t.WiredDatatype
}

// ExecuteOperationWithTransaction is a method to execute an operation with a transaction.
// an operation can be either local or remote
func (t *TransactionDatatype) ExecuteOperationWithTransaction(ctx *TransactionContext, op operations.Operation, isLocal bool) (interface{}, error) {
	transactionCtx, err := t.BeginTransaction(NotUserTransactionTag, ctx, false)
	if err != nil {
		return 0, t.Logger.OrtooErrorf(err, "fail to execute transaction")
	}
	defer func() {
		if err := t.EndTransaction(transactionCtx, false, isLocal); err != nil {
			_ = log.OrtooError(err)
		}
	}()

	if isLocal {
		ret, err := t.executeLocalBase(op)
		if err != nil {
			return 0, t.Logger.OrtooErrorf(err, "fail to execute operation")
		}
		t.currentTrxCtx.appendOperation(op)
		return ret, nil
	}
	t.executeRemoteBase(op)
	return nil, nil
}

// make a transaction and lock
func (t *TransactionDatatype) setTransactionContextAndLock(tag string) *TransactionContext {
	if tag != NotUserTransactionTag {
		t.Logger.Infof("Begin the transaction: `%s`", tag)
	}
	t.mutex.Lock()
	t.isLocked = true
	transactionCtx := &TransactionContext{
		tag:          tag,
		opBuffer:     nil,
		rollbackOpID: t.opID.Clone(),
	}
	return transactionCtx
}

// BeginTransaction is called before a transaction is executed.
// This sets TransactionDatatype.currentTrxCtx, lock, and generates a transaction operation
// This is called in either DoTransaction() or ExecuteOperationWithTransaction().
// Note that TransactionDatatype.currentTrxCtx is currently working transaction context.
func (t *TransactionDatatype) BeginTransaction(tag string, tnxCtx *TransactionContext, withOp bool) (*TransactionContext, error) {
	if t.isLocked && t.currentTrxCtx == tnxCtx {
		return nil, nil // called after DoTransaction() succeeds.
	}
	t.currentTrxCtx = t.setTransactionContextAndLock(tag)
	if withOp {
		op := operations.NewTransactionOperation(tag)
		t.SetNextOpID(op)
		t.currentTrxCtx.appendOperation(op)
	}
	return t.currentTrxCtx, nil
}

// Rollback is called to rollback a transaction
func (t *TransactionDatatype) Rollback() error {
	t.Logger.Infof("Begin the rollback: '%s'", t.currentTrxCtx.tag)
	snapshotDatatype, _ := t.datatype.(SnapshotDatatype)
	redoOpID := t.opID
	redoSnapshot := snapshotDatatype.GetSnapshot().CloneSnapshot()
	t.SetOpID(t.currentTrxCtx.rollbackOpID)
	snapshotDatatype.SetSnapshot(t.rollbackSnapshot)
	for _, op := range t.rollbackOps {
		err := t.Replay(op)
		if err != nil {
			t.SetOpID(redoOpID)
			snapshotDatatype.SetSnapshot(redoSnapshot)
			return t.Logger.OrtooErrorf(err, "fail to replay operations")
		}
	}
	t.rollbackOpID = t.opID.Clone()
	t.rollbackSnapshot = snapshotDatatype.GetSnapshot().CloneSnapshot()
	t.rollbackOps = nil
	t.Logger.Infof("End the rollback: '%s'", t.currentTrxCtx.tag)
	return nil
}

// SetTransactionFail is called when a transaction fails
func (t *TransactionDatatype) SetTransactionFail() {
	t.success = false
}

// EndTransaction is called when a transaction ends
func (t *TransactionDatatype) EndTransaction(trxCtx *TransactionContext, withOp, isLocal bool) error {
	if trxCtx == t.currentTrxCtx {
		defer t.unlock()
		if t.success {
			if withOp {
				beginOp, ok := t.currentTrxCtx.opBuffer[0].(*operations.TransactionOperation)
				if !ok {
					return errors.NewDatatypeError(errors.ErrDatatypeTransaction, "no transaction operation")
				}
				beginOp.SetNumOfOps(len(t.currentTrxCtx.opBuffer))
			}
			t.rollbackOps = append(t.rollbackOps, t.currentTrxCtx.opBuffer...)
			if isLocal {
				t.deliverTransaction(t.currentTrxCtx.opBuffer)
			}
			if t.currentTrxCtx.tag != NotUserTransactionTag {
				t.Logger.Infof("End the transaction: `%s`", t.currentTrxCtx.tag)
			}
		} else {
			if err := t.Rollback(); err != nil {
				panic(err)
			}

		}
	}
	return nil
}

func (t *TransactionDatatype) unlock() {
	if t.isLocked {
		t.currentTrxCtx = nil
		t.success = true
		t.mutex.Unlock()
		t.isLocked = false
	}
}
