package model

import (
	"github.com/knowhunger/ortoo/commons/log"
)

//Operation defines the interfaces of Operation
type Operation interface {
	ExecuteLocal(datatype FinalDatatype) (interface{}, error)
	ExecuteRemote(datatype FinalDatatype) (interface{}, error)
	GetBase() *BaseOperation
}

//NewOperation creates a new operation.
func NewOperation(opType TypeOfOperation) *BaseOperation {
	return &BaseOperation{
		Id:     NewOperationID(),
		OpType: opType,
	}
}

//SetOperationID sets the ID of an operation.
func (o *BaseOperation) SetOperationID(opID *OperationID) {
	o.Id = opID
}

//////////////////// TransactionOperation ////////////////////

//NewTransactionBeginOperation creates a transaction operation
func NewTransactionBeginOperation(tag string) (*TransactionOperation, error) {
	uuid, err := newUniqueID()
	if err != nil {
		return nil, log.OrtooError(err, "fail to create uuid")
	}
	return &TransactionOperation{
		Base: NewOperation(TypeOfOperation_TRANSACTION_BEGIN),
		Uuid: uuid,
		Tag:  tag,
	}, nil
}

//ExecuteLocal ...
func (t *TransactionOperation) ExecuteLocal(datatype FinalDatatype) (interface{}, error) {
	return nil, nil
}

//ExecuteRemote ...
func (t *TransactionOperation) ExecuteRemote(datatype FinalDatatype) (interface{}, error) {
	//datatype.BeginTransaction(t.Tag)
	return nil, nil
}

//////////////////// IncreaseOperation ////////////////////

//NewIncreaseOperation creates a new IncreaseOperation of IntCounter
func NewIncreaseOperation(delta int32) *IncreaseOperation {
	return &IncreaseOperation{
		Base:  NewOperation(TypeOfOperation_INT_COUNTER_INCREASE),
		Delta: delta,
	}
}

//ExecuteLocal ...
func (i *IncreaseOperation) ExecuteLocal(datatype FinalDatatype) (interface{}, error) {
	return datatype.ExecuteLocal(i)
}

//ExecuteRemote ...
func (i *IncreaseOperation) ExecuteRemote(datatype FinalDatatype) (interface{}, error) {
	return datatype.ExecuteRemote(i)
}

//ToOperationOnWire transforms an Operation to OperationOnWire.
func ToOperationOnWire(op Operation) *OperationOnWire {
	switch o := op.(type) {
	case *IncreaseOperation:
		return &OperationOnWire{Body: &OperationOnWire_IncreaseOperation{o}}
	case *TransactionOperation:
		return &OperationOnWire{Body: &OperationOnWire_TransactionOperation{o}}
		//case *TransactionEndOperation:
		//	return &OperationOnWire{Body: &OperationOnWire_TransactionEndOperation{o}}
	}
	return nil
}

//ToOperation transforms an OperationOnWire to Operation.
func ToOperation(op *OperationOnWire) Operation {
	switch o := op.Body.(type) {
	case *OperationOnWire_IncreaseOperation:
		return o.IncreaseOperation
	}
	return nil
}
