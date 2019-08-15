package datatypes

import (
	"github.com/knowhunger/ortoo/commons/model"
	"sync"
)

type TransactionManager struct {
	mutex      *sync.RWMutex
	isLocked   bool
	numOp      uint32
	opBuffer   []model.Operation
	currentCtx *TransactionContext
}

type TransactionContext struct {
	tag string
}

func NewTransactionManager() *TransactionManager {
	return &TransactionManager{
		mutex:      new(sync.RWMutex),
		isLocked:   false,
		numOp:      0,
		currentCtx: nil,
	}
}

func (t *TransactionManager) BeginTransaction(tag string, ctx *TransactionContext) *TransactionContext {
	if t.isLocked && t.currentCtx == ctx { // after called doTransaction
		return nil
	} else {
		t.mutex.Lock()
		t.isLocked = true
		t.currentCtx = &TransactionContext{
			tag: tag,
		}
		return t.currentCtx
	}
}

func (t *TransactionManager) EndTransaction(ctx *TransactionContext) {
	if ctx == t.currentCtx {
		t.mutex.Unlock()
		t.isLocked = false
		t.currentCtx = nil
	}
}

func (t *TransactionManager) Executable(ctx *TransactionContext, op model.Operation) {
	//if ctx == nil {
	//	ctx :=t.BeginTransaction("", ctx)
	//
	//}

}
