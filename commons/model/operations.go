package model

import "github.com/knowhunger/ortoo/commons/log"

type Operationer interface {
	ExecuteLocal(datatype OperationExecuter) (interface{}, error)
	ExecuteRemote(datatype OperationExecuter) (interface{}, error)
	GetBase() *BaseOperation
}

type OperationExecuter interface {
	ExecuteLocal(op interface{}) (interface{}, error)
	ExecuteRemote(op interface{}) (interface{}, error)
}

func NewOperation(opType TypeOperation) *BaseOperation {
	return &BaseOperation{
		Id:     NewOperationID(),
		OpType: opType,
	}
}

func (o *BaseOperation) SetOperationID(opID *OperationID) {
	o.Id = opID
}

//////////////////// TransactionOperation ////////////////////

func NewTransactionBeginOperation() (*TransactionBeginOperation, error) {
	uuid, err := newUniqueID()
	if err != nil {
		return nil, log.OrtooError(err, "fail to create uuid")
	}
	return &TransactionBeginOperation{
		Base: NewOperation(TypeOperation_TRANSACTION_BEGIN),
		Uuid: uuid,
	}, nil
}

func (t *TransactionBeginOperation) ExecuteLocal(datatype OperationExecuter) (interface{}, error) {
	return nil, nil
}

func (t *TransactionBeginOperation) ExecuteRemote(datatype OperationExecuter) (interface{}, error) {
	return nil, nil
}

func NewTransactionEndOperation(uuid uniqueID, numOfOp uint32) *TransactionEndOperation {
	return &TransactionEndOperation{
		Base:     NewOperation(TypeOperation_TRANSACTION_END),
		Uuid:     uuid,
		NumOfOps: numOfOp,
	}
}

func (t *TransactionEndOperation) ExecuteLocal(datatype OperationExecuter) (interface{}, error) {
	return nil, nil
}

func (t *TransactionEndOperation) ExecuteRemote(datatype OperationExecuter) (interface{}, error) {
	return nil, nil
}

//////////////////// IncreaseOperation ////////////////////

func NewIncreaseOperation(delta int32) *IncreaseOperation {
	return &IncreaseOperation{
		Base:  NewOperation(TypeOperation_INT_COUNTER_INCREASE),
		Delta: delta,
	}
}

func (i *IncreaseOperation) ExecuteLocal(datatype OperationExecuter) (interface{}, error) {
	return datatype.ExecuteLocal(i)
}

func (i *IncreaseOperation) ExecuteRemote(datatype OperationExecuter) (interface{}, error) {
	return datatype.ExecuteRemote(i)
}

func ToOperation(op Operationer) *Operation {
	switch o := op.(type) {
	case *IncreaseOperation:
		return &Operation{Body: &Operation_IncreaseOperation{o}}
	case *TransactionBeginOperation:
		return &Operation{Body: &Operation_TransactionBeginOperation{o}}
	case *TransactionEndOperation:
		return &Operation{Body: &Operation_TransactionEndOperation{o}}
	}
	return nil
}

func ToOperationer(op *Operation) Operationer {
	switch o := op.Body.(type) {
	case *Operation_IncreaseOperation:
		return o.IncreaseOperation
	}
	return nil
}
