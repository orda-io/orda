package operations

import (
	"encoding/hex"
	"fmt"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/ortoo/model"
)

// Operation defines the interfaces of Operation
type Operation interface {
	SetOperationID(opID *model.OperationID)
	ExecuteLocal(datatype model.Datatype) (interface{}, error)
	ExecuteRemote(datatype model.Datatype) (interface{}, error)
	ToModelOperation() *model.Operation
	GetType() model.TypeOfOperation
	String() string
	GetID() *model.OperationID
	GetAsJSON() interface{}
}

// ////////////////// BaseOperation ////////////////////

func NewBaseOperation(opID *model.OperationID) *BaseOperation {
	return &BaseOperation{
		ID: opID,
	}
}

type BaseOperation struct {
	ID *model.OperationID
}

func (its *BaseOperation) SetOperationID(opID *model.OperationID) {
	its.ID = opID
}

func (its *BaseOperation) GetID() *model.OperationID {
	return its.ID
}

func (its *BaseOperation) GetAsJSON() interface{} {
	return struct {
		Era     uint32
		Lamport uint64
		CUID    string
		Seq     uint64
	}{
		Era:     its.ID.Era,
		Lamport: its.ID.Lamport,
		CUID:    hex.EncodeToString(its.ID.CUID),
		Seq:     its.ID.Seq,
	}
}

func toString(id *model.OperationID, content interface{}) string {
	return fmt.Sprintf("%s|%v", id.ToString(), content)
}

// ////////////////// TransactionOperation ////////////////////

func NewTransactionOperation(tag string) *TransactionOperation {
	return &TransactionOperation{
		BaseOperation: NewBaseOperation(nil),
		C: TransactionContent{
			Tag: tag,
		},
	}
}

type TransactionContent struct {
	Tag      string
	NumOfOps int32
}

type TransactionOperation struct {
	*BaseOperation
	C TransactionContent
}

func (its *TransactionOperation) GetType() model.TypeOfOperation {
	return model.TypeOfOperation_TRANSACTION
}

func (its *TransactionOperation) ExecuteLocal(datatype model.Datatype) (interface{}, error) {
	return nil, nil
}

func (its *TransactionOperation) ExecuteRemote(datatype model.Datatype) (interface{}, error) {
	return nil, nil
}

func (its *TransactionOperation) String() string {
	return toString(its.ID, its.C)
}

func (its *TransactionOperation) ToModelOperation() *model.Operation {
	return &model.Operation{
		ID:     its.ID,
		OpType: model.TypeOfOperation_TRANSACTION,
		Json:   marshalContent(its.C),
	}
}

func (its *TransactionOperation) SetNumOfOps(numOfOps int) {
	its.C.NumOfOps = int32(numOfOps)
}

func (its *TransactionOperation) GetNumOfOps() int32 {
	return its.C.NumOfOps
}

func (its *TransactionOperation) GetAsJSON() interface{} {
	return &struct {
		ID   interface{}
		Type string
		TransactionContent
	}{
		ID:                 its.BaseOperation.GetAsJSON(),
		Type:               model.TypeOfOperation_TRANSACTION.String(),
		TransactionContent: its.C,
	}
}

// ////////////////// ErrorOperation ////////////////////

func NewErrorOperation(err *model.PushPullError) *ErrorOperation {
	return &ErrorOperation{
		BaseOperation: nil,
		C: ErrorContent{
			Code: int32(err.Code),
			Msg:  err.Msg,
		},
	}
}

type ErrorContent struct {
	Code int32
	Msg  string
}

type ErrorOperation struct {
	*BaseOperation
	C ErrorContent
}

func (its *ErrorOperation) ExecuteLocal(datatype model.Datatype) (interface{}, error) {
	panic("should not be called")
}

func (its *ErrorOperation) ExecuteRemote(datatype model.Datatype) (interface{}, error) {
	panic("should not be called")
}

func (its *ErrorOperation) ToModelOperation() *model.Operation {
	return &model.Operation{
		ID:     its.ID,
		OpType: model.TypeOfOperation_ERROR,
		Json:   marshalContent(its.C),
	}
}

func (its *ErrorOperation) GetType() model.TypeOfOperation {
	return model.TypeOfOperation_ERROR
}

func (its *ErrorOperation) String() string {
	return toString(its.ID, its.C)
}

func (its *ErrorOperation) GetAsJSON() interface{} {
	return &struct {
		ID   interface{}
		Type string
		ErrorContent
	}{
		ID:           its.BaseOperation.GetAsJSON(),
		Type:         model.TypeOfOperation_ERROR.String(),
		ErrorContent: its.C,
	}
}

// GetPushPullError returns PushPullError from ErrorOperation
func (its *ErrorOperation) GetPushPullError() *model.PushPullError {
	return &model.PushPullError{
		Code: model.ErrorCodePushPull(its.C.Code),
		Msg:  its.C.Msg,
	}
}

// ////////////////// SnapshotOperation ////////////////////

func NewSnapshotOperation(typeOf model.TypeOfDatatype, state model.StateOfDatatype, snapshot model.Snapshot) (*SnapshotOperation, error) {
	json, err := snapshot.GetAsJSON()
	if err != nil {
		return nil, log.OrtooError(err)
	}
	return &SnapshotOperation{
		BaseOperation: NewBaseOperation(nil),
		C: SnapshotContent{
			Type:     typeOf,
			State:    state,
			Snapshot: json,
		},
	}, nil
}

type SnapshotOperation struct {
	*BaseOperation
	C SnapshotContent
}

type SnapshotContent struct {
	Type     model.TypeOfDatatype
	State    model.StateOfDatatype
	Snapshot string
}

func (its *SnapshotOperation) ExecuteLocal(datatype model.Datatype) (interface{}, error) {
	datatype.SetState(its.C.State)
	return nil, nil
}

func (its *SnapshotOperation) ExecuteRemote(datatype model.Datatype) (interface{}, error) {
	return datatype.ExecuteRemote(its)
}

func (its *SnapshotOperation) ToModelOperation() *model.Operation {
	return &model.Operation{
		ID:     its.ID,
		OpType: model.TypeOfOperation_SNAPSHOT,
		Json:   marshalContent(its.C),
	}
}

func (its *SnapshotOperation) GetType() model.TypeOfOperation {
	return model.TypeOfOperation_SNAPSHOT
}

func (its *SnapshotOperation) String() string {
	return toString(its.ID, its.C)
}

func (its *SnapshotOperation) GetAsJSON() interface{} {
	return &struct {
		ID   interface{}
		Type string
		SnapshotContent
	}{
		ID:              its.BaseOperation.GetAsJSON(),
		Type:            model.TypeOfOperation_SNAPSHOT.String(),
		SnapshotContent: its.C,
	}
}
