package datatypes

import (
	"bytes"
	"github.com/knowhunger/ortoo/commons/errors"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
	"sync"
)

type TransactionManager struct {
	*WiredDatatypeImpl
	mutex      *sync.RWMutex
	isLocked   bool
	success    bool
	numOp      uint32
	currentCtx *TransactionContext
}

type TransactionContext struct {
	tag      string
	opBuffer []model.Operation
	uuid     []byte
}

func (t *TransactionContext) appendOperation(op model.Operation) {
	t.opBuffer = append(t.opBuffer, op)
}

func NewTransactionManager(ty model.TypeDatatype, w Wire) (*TransactionManager, error) {
	wiredDatatype, err := NewWiredDataType(ty, w)
	if err != nil {
		return nil, log.OrtooError(err, "fail to create int counter due to wiredDatatype")
	}
	return &TransactionManager{
		WiredDatatypeImpl: wiredDatatype,
		mutex:             new(sync.RWMutex),
		isLocked:          false,
		success:           true,
		numOp:             0,
		currentCtx:        nil,
	}, nil
}

func (t *TransactionManager) Execute(ctx *TransactionContext, op model.Operation) (interface{}, error) {
	transactionCtx, err := t.BeginTransaction("", ctx, false)
	if err != nil {
		return 0, log.OrtooError(err, "fail to execute transaction")
	}
	defer t.EndTransaction(transactionCtx, false)
	ret, err := t.executeLocalBase(op)
	t.currentCtx.appendOperation(op)
	if err != nil {
		return 0, log.OrtooError(err, "fail to execute operation")
	}
	return ret.(int32), nil
}

func (t *TransactionManager) BeginTransaction(tag string, ctx *TransactionContext, withOp bool) (*TransactionContext, error) {
	if t.isLocked && t.currentCtx == ctx { // after called doTransaction
		return nil, nil
	} else {
		t.mutex.Lock()
		t.isLocked = true
		t.currentCtx = &TransactionContext{
			tag: tag,
		}
		if withOp {
			op, err := model.NewTransactionBeginOperation(tag)
			if err != nil {
				return nil, log.OrtooError(err, "fail to create TransactionBeginOperation")
			}
			t.currentCtx.uuid = op.Uuid
			t.SetNextOpID(op)
			t.currentCtx.appendOperation(op)
		}
		return t.currentCtx, nil
	}
}

func (t *TransactionManager) SetTransactionFail() {
	t.success = false
}

func (t *TransactionManager) EndTransaction(ctx *TransactionContext, withOp bool) error {
	if ctx == t.currentCtx {
		defer t.unlock()
		if t.success {
			if withOp {
				op := model.NewTransactionEndOperation(t.currentCtx.uuid, uint32(len(t.currentCtx.opBuffer)+1))
				t.SetNextOpID(op)
				t.currentCtx.appendOperation(op)
				err := validateTransaction(t.currentCtx.opBuffer)
				if err != nil {
					return log.OrtooError(err, "fail to validate transaction")
				}
			}
			t.deliverTransaction(t.currentCtx.opBuffer)
		} else {
			t.rollback()
		}
	}
	return nil
}

func (t *TransactionManager) unlock() {
	t.isLocked = false
	t.currentCtx = nil
	t.success = true
	t.mutex.Unlock()
}

func (t *TransactionManager) rollback() {
	log.Logger.Error("not implemented yet")
}

func validateTransaction(transaction []model.Operation) error {
	beginOp, ok := transaction[0].(*model.TransactionBeginOperation)
	if !ok {
		return log.OrtooError(errors.NewTransactionError(), "invalidate transaction: no begin transaction")
	}
	endOp, ok := transaction[len(transaction)-1].(*model.TransactionEndOperation)
	if !ok {
		return log.OrtooError(errors.NewTransactionError(), "invalidate transaction: no end transaction")
	}
	if !bytes.Equal(beginOp.Uuid, endOp.Uuid) {
		return log.OrtooError(errors.NewTransactionError(), "invalidate transaction: not match transaction operations")
	}
	if int(endOp.NumOfOps) != len(transaction) {
		return log.OrtooError(errors.NewTransactionError(), "invalidate transaction: incorrect number of operations")
	}
	beginOp.NumOfOps = endOp.NumOfOps
	return nil
}
