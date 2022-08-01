package operations

import (
	"github.com/orda-io/orda/client/pkg/errors"
	"github.com/orda-io/orda/client/pkg/model"
)

// ////////////////// TransactionOperation ////////////////////

// NewTransactionOperation creates a TransactionOperation.
func NewTransactionOperation(tag string) *TransactionOperation {
	return &TransactionOperation{
		baseOperation: newBaseOperation(
			model.TypeOfOperation_TRANSACTION,
			nil,
			&transactionBody{
				Tag: tag,
			},
		),
	}
}

type transactionBody struct {
	Tag      string
	NumOfOps int32
}

// TransactionOperation is used to begin a transaction.
type TransactionOperation struct {
	baseOperation
}

func (its *TransactionOperation) GetBody() *transactionBody {
	return its.Body.(*transactionBody)
}

// SetNumOfOps sets the number operations in the transaction.
func (its *TransactionOperation) SetNumOfOps(numOfOps int) {
	its.GetBody().NumOfOps = int32(numOfOps)
}

// GetNumOfOps returns the number operations in the transaction.
func (its *TransactionOperation) GetNumOfOps() int32 {
	return its.GetBody().NumOfOps
}

// ////////////////// ErrorOperation ////////////////////

// NewErrorOperation creates an ErrorOperation.
func NewErrorOperation(err errors.OrdaError) *ErrorOperation {
	return &ErrorOperation{
		baseOperation: newBaseOperation(
			model.TypeOfOperation_ERROR,
			model.NewOperationID(),
			&errorBody{
				Code: err.GetCode(),
				Msg:  err.Error(),
			},
		),
	}
}

func NewErrorOperationWithCodeAndMsg(code errors.ErrorCode, msg string) *ErrorOperation {
	return &ErrorOperation{
		baseOperation: newBaseOperation(
			model.TypeOfOperation_ERROR,
			model.NewOperationID(),
			&errorBody{
				Code: code,
				Msg:  msg,
			},
		),
	}
}

type errorBody struct {
	Code errors.ErrorCode
	Msg  string
}

// ErrorOperation is used to deliver an error.
type ErrorOperation struct {
	baseOperation
}

func (its *ErrorOperation) getBody() *errorBody {
	return its.Body.(*errorBody)
}

// GetPushPullError returns PushPullError from ErrorOperation
func (its *ErrorOperation) GetPushPullError() *errors.PushPullError {
	return &errors.PushPullError{
		Code: its.getBody().Code,
		Msg:  its.getBody().Msg,
	}
}

func (its *ErrorOperation) GetCode() errors.ErrorCode {
	return its.Body.(*errorBody).Code
}

func (its *ErrorOperation) GetMessage() string {
	return its.Body.(*errorBody).Msg
}

// ////////////////// SnapshotOperation ////////////////////

// NewSnapshotOperation creates a SnapshotOperation
func NewSnapshotOperation(typeOf model.TypeOfDatatype, snapshot []byte) *SnapshotOperation {
	var typeOfOp = model.TypeOfOperation(typeOf*10 + 10)

	return &SnapshotOperation{
		baseOperation: newBaseOperation(typeOfOp, model.NewOperationID(), snapshot),
	}
}

// SnapshotOperation is used to deliver the snapshot of a datatype.
type SnapshotOperation struct {
	baseOperation
}

func (its *SnapshotOperation) GetBody() []byte {
	return its.Body.([]byte)
}
