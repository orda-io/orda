package errors

import (
	"fmt"
	"github.com/knowhunger/ortoo/ortoo/log"
)

type DatatypeErrorCode ErrorCode

const baseDatatypeCode DatatypeErrorCode = 200

// ErrDatatypeXXX defines an error related to Datatype
const (
	ErrDatatypeCreate = baseDatatypeCode + iota
	ErrDatatypeSubscribe
	ErrDatatypeTransaction
	ErrDatatypeSnapshot
	ErrDatatypeInvalidType
	ErrDatatypeIllegalOperation
	ErrDatatypeInvalidParent
	ErrDatatypeNotExistChildDocument
	ErrDatatypeNoOp
)

var datatypeErrFormats = map[DatatypeErrorCode]string{
	ErrDatatypeCreate:                "fail to create datatype: %s",
	ErrDatatypeSubscribe:             "fail to subscribe datatype: %s",
	ErrDatatypeTransaction:           "fail to proceed transaction: %s",
	ErrDatatypeSnapshot:              "fail to make a snapshot: %s",
	ErrDatatypeInvalidType:           "fail to make an operation due to invalid value type: %s",
	ErrDatatypeIllegalOperation:      "fail to execute operation due to illegal operation: %v",
	ErrDatatypeNotExistChildDocument: "fail to retrieve child due to absence",
	ErrDatatypeInvalidParent:         "fail to access child with invalid parent",
	ErrDatatypeNoOp:                  "fail to issue operation",
}

func (its DatatypeErrorCode) New(args ...interface{}) OrtooError {
	format := fmt.Sprintf("[DatatypeError: %d] %s", its, datatypeErrFormats[its])
	err := &OrtooErrorImpl{
		Code: ErrorCode(its),
		Msg:  fmt.Sprintf(format, args...),
	}
	_ = log.OrtooErrorWithSkip(err, 3, err.Msg)
	return err
}

func (its DatatypeErrorCode) ToErrorCode() ErrorCode {
	return ErrorCode(its)
}
