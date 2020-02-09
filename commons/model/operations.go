package model

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/knowhunger/ortoo/commons/log"
)

// Operation defines the interfaces of Operation
type Operation interface {
	ExecuteLocal(datatype FinalDatatype) (interface{}, error)
	ExecuteRemote(datatype FinalDatatype) (interface{}, error)
	GetBase() *BaseOperation
	ToString() string
	GetAsJson() interface{}
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

func (o *BaseOperation) ToBaseString() string {
	return fmt.Sprintf("%s|%s", o.OpType.String(), o.ID.ToString())
}

func (o *BaseOperation) GetAsJson() interface{} {
	return &struct {
		Era     uint32
		Lamport uint64
		CUID    string
		Seq     uint64
	}{
		Era:     o.ID.Era,
		Lamport: o.ID.Lamport,
		CUID:    hex.EncodeToString(o.ID.CUID),
		Seq:     o.ID.Seq,
	}
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

func (t *TransactionOperation) ToString() string {
	return fmt.Sprintf("%s %s(%s) len:%d", t.Base.ToBaseString(), t.Tag, hex.EncodeToString(t.Uuid), t.NumOfOps)
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

func (s *SnapshotOperation) ToString() string {
	return fmt.Sprintf("%s:%v", s.Base.ToBaseString(), s.Snapshot.Value)
}

func (i *SnapshotOperation) GetAsJson() interface{} {
	val =
	return &struct {
		ID    interface{}
		Type  string
		Value int32
	}{
		ID:    i.Base.GetAsJson(),
		Type:  i.Base.OpType.String(),
		Value: i.Snapshot.Value,
	}
}

// ////////////////// ErrorOperation ////////////////////
func NewErrorOperation(err *PushPullError) *ErrorOperation {
	return &ErrorOperation{
		Base: NewOperation(TypeOfOperation_ERROR),
		Code: uint32(err.Code),
		Msg:  err.Msg,
	}
}

// ExecuteLocal ...
func (e *ErrorOperation) ExecuteLocal(datatype FinalDatatype) (interface{}, error) {
	return nil, nil
}

// ExecuteRemote ...
func (e *ErrorOperation) ExecuteRemote(datatype FinalDatatype) (interface{}, error) {
	return datatype.ExecuteRemote(e)
}

func (e *ErrorOperation) ToString() string {
	return fmt.Sprintf("%s:PushPullErr(%d): %s", e.Base.ToBaseString(), e.Code, e.Msg)
}

func (e *ErrorOperation) GetPushPullErr() *PushPullError {
	return &PushPullError{
		Code: errorCodePushPull(e.Code),
		Msg:  e.Msg,
	}
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

func (i *IncreaseOperation) GetAsJson() interface{} {
	return &struct {
		ID    interface{}
		Type  string
		Value int32
	}{
		ID:    i.Base.GetAsJson(),
		Type:  i.Base.OpType.String(),
		Value: i.Delta,
	}
}

func (i *IncreaseOperation) ToString() string {
	j := i.GetAsJson()
	str, _ := json.Marshal(j)
	return string(str) // fmt.Sprintf("%s delta:%d", i.Base.ToBaseString(), i.Delta)
}

// ToOperationOnWire transforms an Operation to OperationOnWire.
func ToOperationOnWire(op Operation) *OperationOnWire {
	switch o := op.(type) {
	case *SnapshotOperation:
		return &OperationOnWire{Body: &OperationOnWire_Snapshot{o}}
	case *ErrorOperation:
		return &OperationOnWire{Body: &OperationOnWire_Error{o}}
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
	case *OperationOnWire_Error:
		return o.Error
	case *OperationOnWire_Increase:
		return o.Increase
	case *OperationOnWire_Transaction:
		return o.Transaction
	}
	return nil
}

func (o *OperationOnWire) ToString() string {
	return ToOperation(o).ToString()
}
