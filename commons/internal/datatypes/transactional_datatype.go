package datatypes

import (
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
	"sync"
)

type transactionalDatatype struct {
	*baseDatatype
	mutex             *sync.RWMutex
	isLocked          bool
	wg                sync.WaitGroup
	uuid              []byte
	numOp             uint32
	transactionBuffer []model.Operation
	success           bool
}

type PublicTransactionalDatatypeInterface interface {
	PublicBaseDatatypeInterface
}

func newTransactionalDatatype(t model.TypeDatatype) (*transactionalDatatype, error) {
	baseDatatype, err := newBaseDatatype(t)
	if err != nil {
		return nil, log.OrtooError(err, "fail to create transactional datatype")
	}
	return &transactionalDatatype{
		baseDatatype: baseDatatype,
		mutex:        &sync.RWMutex{},
		isLocked:     false,
		success:      true,
	}, nil
}

func (t *transactionalDatatype) executeLocalNotTransactional(datatype model.OperationExecuter, op model.Operation) (interface{}, error) {
	//t.BeginTransaction()
	defer t.EndTransaction()
	return t.executeLocalBase(op)
}

func (t *transactionalDatatype) executeLocalTransactional(datatype model.OperationExecuter, op model.Operation) (interface{}, error) {
	ret, err := t.executeLocalBase(op)
	if err != nil {
		return ret, log.OrtooError(err, "fail to executeLocalBase")
	}

	return nil, nil
}

func (t *transactionalDatatype) BeginTransaction(tag string) {
	t.mutex.Lock()
	t.isLocked = true
	log.Logger.Info("begin transaction:%s", tag)
}

func (t *transactionalDatatype) BeginTransactionLocal(tag string) error {
	t.BeginTransaction(tag)
	op, err := model.NewTransactionBeginOperation(tag)
	if err != nil {
		return log.OrtooError(err, "fail to create new TransactionBeginOperation")
	}
	t.uuid = op.Uuid
	t.numOp = 0
	t.SetNextOpID(op)
	t.success = true
	t.transactionBuffer = append(t.transactionBuffer, op)
	return nil
}

func (t *transactionalDatatype) EndTransaction() {
	t.mutex.Unlock()
	t.isLocked = false
	log.Logger.Info("end transaction")
}

func (t *transactionalDatatype) EndTransactionLocal() []model.Operation {
	if t.success {
		op := model.NewTransactionEndOperation(t.uuid, t.numOp)
		t.SetNextOpID(op)
		t.transactionBuffer = append(t.transactionBuffer, op)
		buffer := t.transactionBuffer
		t.uuid = nil
		t.numOp = 0
		t.transactionBuffer = nil
		return buffer
	}
	return nil
}

func (t *transactionalDatatype) SetTransactionFail() {
	t.success = false
}
