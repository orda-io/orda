package errors

import (
	"fmt"
	"github.com/knowhunger/ortoo/pkg/log"
)

// DatatypeErrorCode is a type for datatype errors
type DatatypeErrorCode ErrorCode

const baseDatatypeCode DatatypeErrorCode = 200

// ErrDatatypeXXX defines an error related to Datatype
const (
	ErrDatatypeCreate = baseDatatypeCode + iota
	ErrDatatypeSubscribe
	ErrDatatypeTransaction
	ErrDatatypeTransactionRollback
	ErrDatatypeSnapshot
	ErrDatatypeIllegalParameters
	ErrDatatypeInvalidParent
	ErrDatatypeNoOp
	ErrDatatypeInvalidValue
	ErrDatatypeMarshal
	ErrDatatypeUnmarshal
	ErrDatatypeNoTarget
)

var datatypeErrFormats = map[DatatypeErrorCode]string{
	ErrDatatypeCreate:            "fail to create datatype: %s",
	ErrDatatypeSubscribe:         "fail to subscribe datatype: %s",
	ErrDatatypeTransaction:       "fail to proceed transaction: %s",
	ErrDatatypeSnapshot:          "fail to make a snapshot: %s",
	ErrDatatypeIllegalParameters: "fail to execute the operation due to illegal operation: %v",
	ErrDatatypeInvalidParent:     "fail to modify due to the invalid parent: %v",
	ErrDatatypeNoOp:              "fail to issue operation: %v",
	ErrDatatypeMarshal:           "fail to marshal:%v",
	ErrDatatypeInvalidValue:      "fail to use the value:%v",
	ErrDatatypeUnmarshal:         "fail to unmarshal:%v",
	ErrDatatypeNoTarget:          "fail to find target: %v",
}

// New creates an error related to the datatype
func (its DatatypeErrorCode) New(l *log.OrtooLog, args ...interface{}) OrtooError {
	format := fmt.Sprintf("[DatatypeError: %d] %s", its, datatypeErrFormats[its])
	err := &singleOrtooError{
		Code: ErrorCode(its),
		Msg:  fmt.Sprintf(format, args...),
	}
	if l != nil {
		_ = l.OrtooSkipErrorf(err, 2, err.Msg)
	} else {
		_ = log.OrtooErrorWithSkip(err, 2, err.Msg)
	}
	return err
}

// ToErrorCode casts DatatypeErrorCode to ErrorCode
func (its DatatypeErrorCode) ToErrorCode() ErrorCode {
	return ErrorCode(its)
}
