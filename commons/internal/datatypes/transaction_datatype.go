package datatypes

import (
	"github.com/knowhunger/ortoo/commons/errors"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
	"sync"
)

type TransactionDatatypeImpl struct {
	*WiredDatatypeImpl
	mutex            *sync.RWMutex
	isLocked         bool
	success          bool
	rollbackSnapshot Snapshot
	rollbackOps      []model.Operation
	rollbackOpID     *model.OperationID
	transactionCtx   *TransactionContext
}

type TransactionDatatype interface {
	ExecuteTransactionRemote(transaction []model.Operation) error
}

type TransactionContext struct {
	tag          string
	opBuffer     []model.Operation
	uuid         []byte
	rollbackOpID *model.OperationID
}

func (t *TransactionContext) GetOpId() *model.OperationID {
	if len(t.opBuffer) > 0 {
		return t.opBuffer[0].GetBase().Id
	}
	return nil
}

func (t *TransactionContext) appendOperation(op model.Operation) {
	t.opBuffer = append(t.opBuffer, op)
}

func NewTransactionManager(ty model.TypeDatatype, w Wire, snapshot Snapshot) (*TransactionDatatypeImpl, error) {
	wiredDatatype, err := NewWiredDataType(ty, w)
	if err != nil {
		return nil, log.OrtooError(err, "fail to create int counter due to wiredDatatype")
	}
	return &TransactionDatatypeImpl{
		WiredDatatypeImpl: wiredDatatype,
		mutex:             new(sync.RWMutex),
		isLocked:          false,
		success:           true,
		transactionCtx:    nil,
		rollbackSnapshot:  snapshot.CloneSnapshot(),
		rollbackOps:       nil,
		rollbackOpID:      wiredDatatype.opID.Clone(),
	}, nil
}

func (t *TransactionDatatypeImpl) ExecuteTransactionRemote(transaction []model.Operation) error {
	var transactionCtx *TransactionContext = nil
	if len(transaction) > 1 {
		if err := validateTransaction(transaction); err != nil {
			return log.OrtooError(err, "fail to validate transaction")
		}
		beginOp := transaction[0].(*model.TransactionOperation)
		transactionCtx = t.beginTransaction(beginOp.Tag)
		defer t.EndTransaction(transactionCtx, false)
	}
	for _, op := range transaction {
		t.ExecuteTransaction(transactionCtx, op, false)
	}
	return nil
}

func (t *TransactionDatatypeImpl) ExecuteTransaction(ctx *TransactionContext, op model.Operation, isLocal bool) (interface{}, error) {
	transactionCtx, err := t.BeginTransaction("", ctx, false)
	if err != nil {
		return 0, log.OrtooError(err, "fail to execute transaction")
	}
	defer t.EndTransaction(transactionCtx, false)
	if isLocal {
		ret, err := t.executeLocalBase(op)
		if err != nil {
			return 0, log.OrtooError(err, "fail to execute operation")
		}
		t.transactionCtx.appendOperation(op)
		return ret.(int32), nil
	} else {
		t.executeRemoteBase(op)
	}
	return nil, nil
}

func (t *TransactionDatatypeImpl) beginTransaction(tag string) *TransactionContext {
	t.mutex.Lock()
	log.Logger.Infof("Begin the transaction: `%s`", tag)
	t.isLocked = true
	t.transactionCtx = &TransactionContext{
		tag:          tag,
		opBuffer:     nil,
		rollbackOpID: t.opID.Clone(),
	}
	return t.transactionCtx
}

func (t *TransactionDatatypeImpl) BeginTransaction(tag string, ctx *TransactionContext, withOp bool) (*TransactionContext, error) {
	if t.isLocked && t.transactionCtx == ctx { // after called doTransaction
		return nil, nil
	} else {
		t.transactionCtx = t.beginTransaction(tag)
		if withOp {
			op, err := model.NewTransactionBeginOperation(tag)
			if err != nil {
				return nil, log.OrtooError(err, "fail to create TransactionBeginOperation")
			}
			t.transactionCtx.uuid = op.Uuid
			t.SetNextOpID(op)
			t.transactionCtx.appendOperation(op)
		}
		return t.transactionCtx, nil
	}
}

func (t *TransactionDatatypeImpl) Rollback() error {
	snapshotDatatype, _ := t.opExecuter.(SnapshotDatatype)
	redoOpID := t.GetBase().opID
	redoSnapshot := snapshotDatatype.GetSnapshot().CloneSnapshot()
	t.SetOpID(t.transactionCtx.rollbackOpID)
	snapshotDatatype.SetSnapshot(t.rollbackSnapshot)
	for _, op := range t.rollbackOps {
		err := t.Replay(op)
		if err != nil {
			t.SetOpID(redoOpID)
			snapshotDatatype.SetSnapshot(redoSnapshot)
			return log.OrtooError(err, "fail to replay operations")
		}
	}
	t.rollbackOpID = t.GetBase().opID.Clone()
	t.rollbackSnapshot = snapshotDatatype.GetSnapshot().CloneSnapshot()
	t.rollbackOps = nil
	return nil
}

func (t *TransactionDatatypeImpl) SetTransactionFail() {
	t.success = false
}

func (t *TransactionDatatypeImpl) EndTransaction(ctx *TransactionContext, withOp bool) error {
	if ctx == t.transactionCtx {
		defer t.unlock()
		if t.success {
			if withOp {
				beginOp, ok := t.transactionCtx.opBuffer[0].(*model.TransactionOperation)
				if !ok {
					return log.OrtooError(errors.NewTransactionError(), "invalidate transaction: no begin operation")
				}
				beginOp.NumOfOps = uint32(len(t.transactionCtx.opBuffer))
			}
			t.rollbackOps = append(t.rollbackOps, t.transactionCtx.opBuffer...)
			t.deliverTransaction(t.transactionCtx.opBuffer)
		} else {
			t.Rollback()
		}
	}
	return nil
}

func (t *TransactionDatatypeImpl) unlock() {
	t.isLocked = false
	t.transactionCtx = nil
	t.success = true
	t.mutex.Unlock()
}

func validateTransaction(transaction []model.Operation) error {
	beginOp, ok := transaction[0].(*model.TransactionOperation)
	if !ok {
		return log.OrtooError(errors.NewTransactionError(), "invalidate transaction: no begin transaction")
	}
	if int(beginOp.NumOfOps) != len(transaction) {
		return log.OrtooError(errors.NewTransactionError(), "invalidate transaction: incorrect number of operations")
	}
	return nil
}
