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

func NewOperation(opType OpType) *Operation {
	return &Operation{
		Id:     NewOperationID(),
		OpType: uint32(opType),
	}
}

func (o *Operation) SetOperationID(opID *OperationID) {
	o.Id = opID
}

func (o *Operation) GetOperationID() *OperationID {
	return o.GetId()
}

//////////////////// IncreaseOperation ////////////////////

func NewIncreaseOperation(delta int32) *IncreaseOperation {
	return &IncreaseOperation{
		Base:  NewOperation(OperationTypes.IntCounterIncreaseType),
		Delta: delta,
	}
}

func (i *IncreaseOperation) ExecuteLocal(datatype OperationExecuter) (interface{}, error) {
	return datatype.ExecuteLocal(i)
}

func (i *IncreaseOperation) ExecuteRemote(datatype OperationExecuter) (interface{}, error) {
	return datatype.ExecuteRemote(i)
}
