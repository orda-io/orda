package operations

import (
	"fmt"
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/iface"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/pkg/types"
)

// ////////////////// baseOperation ////////////////////

func newBaseOperation(opID *model.OperationID) *baseOperation {
	return &baseOperation{
		ID: opID,
	}
}

type baseOperation struct {
	ID *model.OperationID
}

func (its *baseOperation) SetOperationID(opID *model.OperationID) {
	its.ID = opID
}

func (its *baseOperation) GetID() *model.OperationID {
	return its.ID
}

func (its *baseOperation) GetTimestamp() *model.Timestamp {
	return its.ID.GetTimestamp()
}

// GetAsJSON returns the operation in the format of JSON compatible struct.
func (its *baseOperation) GetAsJSON() interface{} {
	return struct {
		Era     uint32
		Lamport uint64
		CUID    string
		Seq     uint64
	}{
		Era:     its.ID.Era,
		Lamport: its.ID.Lamport,
		CUID:    types.UIDtoString(its.ID.CUID),
		Seq:     its.ID.Seq,
	}
}

func toString(id *model.OperationID, content interface{}) string {
	return fmt.Sprintf("%s|%v", id.ToString(), content)
}

// ////////////////// TransactionOperation ////////////////////

// NewTransactionOperation creates a TransactionOperation.
func NewTransactionOperation(tag string) *TransactionOperation {
	return &TransactionOperation{
		baseOperation: newBaseOperation(nil),
		C: transactionContent{
			Tag: tag,
		},
	}
}

type transactionContent struct {
	Tag      string
	NumOfOps int32
}

// TransactionOperation is used to begin a transaction.
type TransactionOperation struct {
	*baseOperation
	C transactionContent
}

// GetType returns the type of operation.
func (its *TransactionOperation) GetType() model.TypeOfOperation {
	return model.TypeOfOperation_TRANSACTION
}

// ExecuteLocal enables the operation to perform something at the local client.
func (its *TransactionOperation) ExecuteLocal(datatype iface.Datatype) (interface{}, errors.OrtooError) {
	return nil, nil
}

// ExecuteRemote enables the operation to perform something at the remote clients.
func (its *TransactionOperation) ExecuteRemote(datatype iface.Datatype) (interface{}, errors.OrtooError) {
	return nil, nil
}

// ToModelOperation transforms this operation to the model.Operation.
func (its *TransactionOperation) String() string {
	return toString(its.ID, its.C)
}

// ToModelOperation transforms this operation to the model.Operation.
func (its *TransactionOperation) ToModelOperation() *model.Operation {
	return &model.Operation{
		ID:     its.ID,
		OpType: model.TypeOfOperation_TRANSACTION,
		Body:   marshalContent(its.C),
	}
}

// SetNumOfOps sets the number operations in the transaction.
func (its *TransactionOperation) SetNumOfOps(numOfOps int) {
	its.C.NumOfOps = int32(numOfOps)
}

// GetNumOfOps returns the number operations in the transaction.
func (its *TransactionOperation) GetNumOfOps() int32 {
	return its.C.NumOfOps
}

// GetAsJSON returns the operation in the format of JSON compatible struct.
func (its *TransactionOperation) GetAsJSON() interface{} {
	return struct {
		ID   interface{}
		Type string
		transactionContent
	}{
		ID:                 its.baseOperation.GetAsJSON(),
		Type:               model.TypeOfOperation_TRANSACTION.String(),
		transactionContent: its.C,
	}
}

// ////////////////// ErrorOperation ////////////////////

// NewErrorOperation creates an ErrorOperation.
func NewErrorOperation(err errors.OrtooError) *ErrorOperation {
	return &ErrorOperation{
		baseOperation: newBaseOperation(model.NewOperationID()),
		C: errorContent{
			Code: int32(err.GetCode()),
			Msg:  err.Error(),
		},
	}
}

type errorContent struct {
	Code int32
	Msg  string
}

// ErrorOperation is used to deliver an error.
type ErrorOperation struct {
	*baseOperation
	C errorContent
}

// ExecuteLocal enables the operation to perform something at the local client.
func (its *ErrorOperation) ExecuteLocal(datatype iface.Datatype) (interface{}, errors.OrtooError) {
	panic("should not be called")
}

// ExecuteRemote enables the operation to perform something at the remote clients.
func (its *ErrorOperation) ExecuteRemote(datatype iface.Datatype) (interface{}, errors.OrtooError) {
	panic("should not be called")
}

// ToModelOperation transforms this operation to the model.Operation.
func (its *ErrorOperation) ToModelOperation() *model.Operation {
	return &model.Operation{
		ID:     its.ID,
		OpType: model.TypeOfOperation_ERROR,
		Body:   marshalContent(its.C),
	}
}

// GetType returns the type of operation.
func (its *ErrorOperation) GetType() model.TypeOfOperation {
	return model.TypeOfOperation_ERROR
}

func (its *ErrorOperation) String() string {
	return toString(its.ID, its.C)
}

// GetAsJSON returns the operation in the format of JSON compatible struct.
func (its *ErrorOperation) GetAsJSON() interface{} {
	return struct {
		ID   interface{}
		Type string
		errorContent
	}{
		ID:           its.baseOperation.GetAsJSON(),
		Type:         model.TypeOfOperation_ERROR.String(),
		errorContent: its.C,
	}
}

// GetPushPullError returns PushPullError from ErrorOperation
func (its *ErrorOperation) GetPushPullError() *errors.PushPullError {
	return &errors.PushPullError{
		Code: errors.PushPullErrorCode(its.C.Code),
		Msg:  its.C.Msg,
	}
}

// ////////////////// SnapshotOperation ////////////////////

// NewSnapshotOperation creates a SnapshotOperation
func NewSnapshotOperation(typeOf model.TypeOfDatatype, state model.StateOfDatatype, snapshot string) *SnapshotOperation {

	return &SnapshotOperation{
		baseOperation: newBaseOperation(nil),
		C: &snapshotContent{
			// Type:     typeOf,
			State:    state,
			Snapshot: snapshot,
		},
	}
}

type snapshotContent struct {
	// Type     model.TypeOfDatatype
	State    model.StateOfDatatype
	Snapshot string
}

func (its *snapshotContent) GetS() string {
	return its.Snapshot
}

// SnapshotOperation is used to deliver the snapshot of a datatype.
type SnapshotOperation struct {
	*baseOperation
	C *snapshotContent
}

// ExecuteLocal enables the operation to perform something at the local client.
func (its *SnapshotOperation) ExecuteLocal(datatype iface.Datatype) (interface{}, errors.OrtooError) {
	datatype.SetState(its.C.State)
	return nil, nil
}

// ExecuteRemote enables the operation to perform something at the remote clients.
func (its *SnapshotOperation) ExecuteRemote(datatype iface.Datatype) (interface{}, errors.OrtooError) {
	return datatype.ExecuteRemote(its)
}

// ToModelOperation transforms this operation to the model.Operation.
func (its *SnapshotOperation) ToModelOperation() *model.Operation {
	return &model.Operation{
		ID:     its.ID,
		OpType: model.TypeOfOperation_SNAPSHOT,
		Body:   marshalContent(its.C),
	}
}

func (its *SnapshotOperation) GetContent() iface.SnapshotContent {
	return its.C
}

// GetType returns the type of operation.
func (its *SnapshotOperation) GetType() model.TypeOfOperation {
	return model.TypeOfOperation_SNAPSHOT
}

func (its *SnapshotOperation) String() string {
	return toString(its.ID, its.C)
}

// GetAsJSON returns the operation in the format of JSON compatible struct.
func (its *SnapshotOperation) GetAsJSON() interface{} {
	return struct {
		ID   interface{}
		Type string
		*snapshotContent
	}{
		ID:              its.baseOperation.GetAsJSON(),
		Type:            model.TypeOfOperation_SNAPSHOT.String(),
		snapshotContent: its.C,
	}
}
