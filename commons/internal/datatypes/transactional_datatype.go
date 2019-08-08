package datatypes

import (
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
	"sync"
)

type transactionalDatatypeImpl struct {
	*baseDatatypeImpl
	mutex    *sync.RWMutex
	isLocked bool
	wg       sync.WaitGroup
	uuid     []byte
	numOp    uint32
}

type TransactionalDatatype interface {
}

type PublicTransactionalInterface interface {
	PublicBaseInterface
	//DoTransaction(func(datatype interface{}) error)
}

func newTransactionalDatatype(t model.TypeDatatype) (*transactionalDatatypeImpl, error) {
	baseDatatype, err := newBaseDatatype(t)
	if err != nil {
		return nil, log.OrtooError(err, "fail to create transactional datatype")
	}
	return &transactionalDatatypeImpl{
		baseDatatypeImpl: baseDatatype,
		mutex:            &sync.RWMutex{},
		isLocked:         false,
	}, nil
}

func (t *transactionalDatatypeImpl) executeLocalTransactional(datatype model.OperationExecuter, op model.Operationer) (interface{}, error) {
	if t.isLocked {

	} else {
		ret, err := t.executeLocalBase(datatype, op)
		if err != nil {
			return ret, log.OrtooError(err, "fail to executeLocalBase")
		}
	}
	return nil, nil
}

func (t *transactionalDatatypeImpl) BeginTransaction() (*model.TransactionBeginOperation, error) {
	// lock
	t.mutex.Lock()
	t.isLocked = true
	op, err := model.NewTransactionBeginOperation()
	if err != nil {
		return nil, log.OrtooError(err, "fail to create new TransactionBeginOperation")
	}
	t.uuid = op.Uuid
	t.numOp = 0
	return op, nil
}

func (t *transactionalDatatypeImpl) EndTransaction() *model.TransactionEndOperation {
	t.mutex.Unlock()
	t.isLocked = false
	numOp := t.numOp
	t.numOp = 0
	return model.NewTransactionEndOperation(t.uuid, numOp)

}
