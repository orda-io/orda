package operations

import (
	errors2 "github.com/orda-io/orda/client/pkg/errors"
	model2 "github.com/orda-io/orda/client/pkg/model"
)

// ////////////////// TransactionOperation ////////////////////

// NewTransactionOperation creates a TransactionOperation.
func NewTransactionOperation(tag string) *TransactionOperation {
	return &TransactionOperation{
		baseOperation: newBaseOperation(
			model2.TypeOfOperation_TRANSACTION,
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
func NewErrorOperation(err errors2.OrdaError) *ErrorOperation {
	return &ErrorOperation{
		baseOperation: newBaseOperation(
			model2.TypeOfOperation_ERROR,
			model2.NewOperationID(),
			&errorBody{
				Code: err.GetCode(),
				Msg:  err.Error(),
			},
		),
	}
}

func NewErrorOperationWithCodeAndMsg(code errors2.ErrorCode, msg string) *ErrorOperation {
	return &ErrorOperation{
		baseOperation: newBaseOperation(
			model2.TypeOfOperation_ERROR,
			model2.NewOperationID(),
			&errorBody{
				Code: code,
				Msg:  msg,
			},
		),
	}
}

type errorBody struct {
	Code errors2.ErrorCode
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
func (its *ErrorOperation) GetPushPullError() *errors2.PushPullError {
	return &errors2.PushPullError{
		Code: its.getBody().Code,
		Msg:  its.getBody().Msg,
	}
}

func (its *ErrorOperation) GetCode() errors2.ErrorCode {
	return its.Body.(*errorBody).Code
}

func (its *ErrorOperation) GetMessage() string {
	return its.Body.(*errorBody).Msg
}

// ////////////////// SnapshotOperation ////////////////////

// NewSnapshotOperation creates a SnapshotOperation
func NewSnapshotOperation(typeOf model2.TypeOfDatatype, snapshot []byte) *SnapshotOperation {
	var typeOfOp = model2.TypeOfOperation(typeOf*10 + 10)

	return &SnapshotOperation{
		baseOperation: newBaseOperation(typeOfOp, model2.NewOperationID(), snapshot),
	}
}

// SnapshotOperation is used to deliver the snapshot of a datatype.
type SnapshotOperation struct {
	baseOperation
}

func (its *SnapshotOperation) GetBody() []byte {
	return its.Body.([]byte)
}
