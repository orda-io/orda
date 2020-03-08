package model

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/knowhunger/ortoo/ortoo/log"
)

// Operation defines the interfaces of Operation
type Operation interface {
	ExecuteLocal(datatype Datatype) (interface{}, error)
	ExecuteRemote(datatype Datatype) (interface{}, error)
	GetBase() *BaseOperation
	ToString() string
	GetAsJSON() interface{}
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

func (o *BaseOperation) GetTimestamp() *Timestamp {
	return o.ID.GetTimestamp()
}

// ToBaseString returns the string for BaseOperation
func (o *BaseOperation) ToBaseString() string {
	return fmt.Sprintf("%s|%s", o.OpType.String(), o.ID.ToString())
}

// GetAsJSON returns the operation as interface{} for JSON
func (o *BaseOperation) GetAsJSON() interface{} {
	return struct {
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

// GetAsJSON returns the operation as interface{} for JSON
func (t *TransactionOperation) GetAsJSON() interface{} {
	return &struct {
		ID   interface{}
		Type string
		Tag  string
	}{
		ID:   t.Base.GetAsJSON(),
		Type: t.Base.OpType.String(),
		Tag:  t.Tag,
	}
}

// ExecuteLocal ...
func (t *TransactionOperation) ExecuteLocal(datatype Datatype) (interface{}, error) {
	return nil, nil
}

// ExecuteRemote ...
func (t *TransactionOperation) ExecuteRemote(datatype Datatype) (interface{}, error) {
	return nil, nil
}

// ToString returns customized string
func (t *TransactionOperation) ToString() string {
	j := t.GetAsJSON()
	str, _ := json.Marshal(j)
	return string(str)
}

// ////////////////// SubscribeOperation ////////////////////

// NewSnapshotOperation generates a new SnapshotOperation
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
func (s *SnapshotOperation) ExecuteLocal(datatype Datatype) (interface{}, error) {
	datatype.SetState(s.State)
	return nil, nil
}

// ExecuteRemote ...
func (s *SnapshotOperation) ExecuteRemote(datatype Datatype) (interface{}, error) {
	return datatype.ExecuteRemote(s)
}

// ToString returns customized string
func (s *SnapshotOperation) ToString() string {
	d, _ := json.Marshal(s.GetAsJSON())
	return string(d)
}

// GetAsJSON returns the operation as interface{} for JSON
func (s *SnapshotOperation) GetAsJSON() interface{} {

	return &struct {
		ID    interface{}
		Type  string
		Value interface{}
	}{
		ID:    s.Base.GetAsJSON(),
		Type:  s.Base.OpType.String(),
		Value: s.Snapshot.Value,
	}
}

// ////////////////// ErrorOperation ////////////////////

// NewErrorOperation generates a new ErrorOperation
func NewErrorOperation(err *PushPullError) *ErrorOperation {
	return &ErrorOperation{
		Base: NewOperation(TypeOfOperation_ERROR),
		Code: uint32(err.Code),
		Msg:  err.Msg,
	}
}

// ExecuteLocal ...
func (e *ErrorOperation) ExecuteLocal(datatype Datatype) (interface{}, error) {
	return nil, nil
}

// ExecuteRemote ...
func (e *ErrorOperation) ExecuteRemote(datatype Datatype) (interface{}, error) {
	return datatype.ExecuteRemote(e)
}

// GetAsJSON returns the operation as interface{} for JSON
func (e *ErrorOperation) GetAsJSON() interface{} {
	return &struct {
		ID   interface{}
		Code uint32
		Msg  string
	}{
		ID:   e.Base.GetAsJSON(),
		Code: e.Code,
		Msg:  e.Msg,
	}
}

// ToString returns customized string
func (e *ErrorOperation) ToString() string {
	data, _ := json.Marshal(e.GetAsJSON())
	return string(data)
}

// GetPushPullError returns PushPullError from ErrorOperation
func (e *ErrorOperation) GetPushPullError() *PushPullError {
	return &PushPullError{
		Code: errorCodePushPull(e.Code),
		Msg:  e.Msg,
	}
}

// ////////////////// PutOperation ////////////////////
func NewPutOperation(key string, value OrtooType) *PutOperation {
	return &PutOperation{
		Base:  NewOperation(TypeOfOperation_HASH_MAP_PUT),
		Key:   key,
		Value: value.(string),
	}
}

func (p *PutOperation) ExecuteLocal(datatype Datatype) (interface{}, error) {
	return datatype.ExecuteLocal(p)
}

func (p *PutOperation) ExecuteRemote(datatype Datatype) (interface{}, error) {
	return datatype.ExecuteRemote(p)
}

// GetAsJSON returns the operation as interface{} for JSON
func (p *PutOperation) GetAsJSON() interface{} {
	return &struct {
		ID    interface{}
		Type  string
		Key   string
		Value interface{}
	}{
		ID:    p.Base.GetAsJSON(),
		Type:  p.Base.OpType.String(),
		Key:   p.Key,
		Value: p.Value,
	}
}

// ToString returns customized string
func (p *PutOperation) ToString() string {
	j := p.GetAsJSON()
	str, _ := json.Marshal(j)
	return string(str)
}

// ////////////////// RemoveOperation ////////////////////
func NewRemoveOperation(key string) *RemoveOperation {
	return &RemoveOperation{
		Base: NewOperation(TypeOfOperation_HASH_MAP_REMOVE),
		Key:  key,
	}
}

func (r *RemoveOperation) ExecuteLocal(datatype Datatype) (interface{}, error) {
	return datatype.ExecuteLocal(r)
}

func (r *RemoveOperation) ExecuteRemote(datatype Datatype) (interface{}, error) {
	return datatype.ExecuteRemote(r)
}

func (r *RemoveOperation) GetAsJSON() interface{} {
	return &struct {
		ID   interface{}
		Type string
		Key  string
	}{
		ID:   r.Base.GetAsJSON(),
		Type: r.Base.OpType.String(),
		Key:  r.Key,
	}
}

func (r *RemoveOperation) ToString() string {
	j := r.GetAsJSON()
	str, _ := json.Marshal(j)
	return string(str)
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
func (i *IncreaseOperation) ExecuteLocal(datatype Datatype) (interface{}, error) {
	return datatype.ExecuteLocal(i)
}

// ExecuteRemote ...
func (i *IncreaseOperation) ExecuteRemote(datatype Datatype) (interface{}, error) {
	return datatype.ExecuteRemote(i)
}

// GetAsJSON returns the operation as interface{} for JSON
func (i *IncreaseOperation) GetAsJSON() interface{} {
	return &struct {
		ID    interface{}
		Type  string
		Value int32
	}{
		ID:    i.Base.GetAsJSON(),
		Type:  i.Base.OpType.String(),
		Value: i.Delta,
	}
}

// ToString returns customized string
func (i *IncreaseOperation) ToString() string {
	j := i.GetAsJSON()
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
	case *PutOperation:
		return &OperationOnWire{Body: &OperationOnWire_Put{o}}
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

// ToString returns customized string
func (o *OperationOnWire) ToString() string {
	return ToOperation(o).ToString()
}
