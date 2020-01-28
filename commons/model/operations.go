package model

import (
	"github.com/knowhunger/ortoo/commons/log"
)

// Operation defines the interfaces of Operation
type Operation interface {
	ExecuteLocal(datatype FinalDatatype) (interface{}, error)
	ExecuteRemote(datatype FinalDatatype) (interface{}, error)
	GetBase() *BaseOperation
}

// NewOperation creates a new operation.
func NewOperation(opType TypeOfOperation) *BaseOperation {
	return &BaseOperation{
		ID:     NewOperationID(),
		OpType: opType,
	}
}

// SetOperationID sets the ID of an operation.
func (o *BaseOperation) SetOperationID(opID *OperationID) {
	o.ID = opID
}

// ////////////////// TransactionOperation ////////////////////

// NewTransactionOperation creates a transaction operation
func NewTransactionOperation(tag string) (*TransactionOperation, error) {
	uuid, err := newUniqueID()
	if err != nil {
		return nil, log.OrtooErrorf(err, "fail to create uuid")
	}
	return &TransactionOperation{
		Base: NewOperation(TypeOfOperation_TRANSACTION),
		Uuid: uuid,
		Tag:  tag,
	}, nil
}

// ExecuteLocal ...
func (t *TransactionOperation) ExecuteLocal(datatype FinalDatatype) (interface{}, error) {
	return nil, nil
}

// ExecuteRemote ...
func (t *TransactionOperation) ExecuteRemote(datatype FinalDatatype) (interface{}, error) {
	// datatype.BeginTransaction(t.Tag)
	return nil, nil
}

// ////////////////// SubscribeOperation ////////////////////
func NewSnapshotOperation(datatype TypeOfDatatype, state StateOfDatatype, snapshot Snapshot) (*SnapshotOperation, error) {
	any, err := snapshot.GetTypeAny()
	if err != nil {
		return nil, log.OrtooErrorf(err, "fail to create subscribe operation")
	}
	return &SnapshotOperation{
		Base:     NewOperation(TypeOfOperation_SNAPSHOT),
		Type:     datatype,
		State:    state,
		Snapshot: any,
	}, nil
}

// ExecuteLocal ...
func (s *SnapshotOperation) ExecuteLocal(datatype FinalDatatype) (interface{}, error) {
	datatype.SetState(s.State)
	return nil, nil
}

// ExecuteRemote ...
func (s *SnapshotOperation) ExecuteRemote(datatype FinalDatatype) (interface{}, error) {

	return datatype.ExecuteRemote(s)
}

// ////////////////// IncreaseOperation ////////////////////

// NewIncreaseOperation creates a new IncreaseOperation of IntCounter
func NewIncreaseOperation(delta int32) *IncreaseOperation {
	return &IncreaseOperation{
		Base:  NewOperation(TypeOfOperation_INT_COUNTER_INCREASE),
		Delta: delta,
	}
}

// ExecuteLocal ...
func (i *IncreaseOperation) ExecuteLocal(datatype FinalDatatype) (interface{}, error) {
	return datatype.ExecuteLocal(i)
}

// ExecuteRemote ...
func (i *IncreaseOperation) ExecuteRemote(datatype FinalDatatype) (interface{}, error) {
	return datatype.ExecuteRemote(i)
}

// ToOperationOnWire transforms an Operation to OperationOnWire.
func ToOperationOnWire(op Operation) *OperationOnWire {
	switch o := op.(type) {
	case *SnapshotOperation:
		return &OperationOnWire{Body: &OperationOnWire_Snapshot{o}}
	case *IncreaseOperation:
		return &OperationOnWire{Body: &OperationOnWire_Increase{o}}
	case *TransactionOperation:
		return &OperationOnWire{Body: &OperationOnWire_Transaction{o}}

	}
	return nil
}

// ToOperation transforms an OperationOnWire to Operation.
func ToOperation(op *OperationOnWire) Operation {
	switch o := op.Body.(type) {
	case *OperationOnWire_Snapshot:
		return o.Snapshot
	case *OperationOnWire_Increase:
		return o.Increase
	case *OperationOnWire_Transaction:
		return o.Transaction
	}
	return nil
}
