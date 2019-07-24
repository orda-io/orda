package model

type Operationer interface {
	ExecuteLocal(datatype OperationExecuter) (interface{}, error)
	ExecuteRemote(datatype OperationExecuter) (interface{}, error)
	GetBase() *Operation
}

type OperationExecuter interface {
	ExecuteLocal(op interface{}) (interface{}, error)
	ExecuteRemote(op interface{}) (interface{}, error)
}

func NewOperation(opType TypeOperation) *Operation {
	return &Operation{
		Id:     NewOperationID(),
		OpType: opType,
	}
}

func (o *Operation) SetOperationID(opID *OperationID) {
	o.Id = opID
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
