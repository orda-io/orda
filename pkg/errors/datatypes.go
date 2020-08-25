package errors

import (
	"fmt"
	"github.com/knowhunger/ortoo/pkg/log"
)

type DatatypeErrorCode ErrorCode

const baseDatatypeCode DatatypeErrorCode = 200

// ErrDatatypeXXX defines an error related to Datatype
const (
	ErrDatatypeCreate = baseDatatypeCode + iota
	ErrDatatypeSubscribe
	ErrDatatypeTransaction
	ErrDatatypeTransactionRollback
	ErrDatatypeSnapshot
	ErrDatatypeInvalidType
	ErrDatatypeIllegalOperation
	ErrDatatypeInvalidParent
	ErrDatatypeNotExistChildDocument
	ErrDatatypeNoOp
	ErrDatatypeMarshal
	ErrDatatypeUnmarshal
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
	ErrDatatypeMarshal:               "fail to marshal:%v",
	ErrDatatypeUnmarshal:             "fail to unmarshal:%v",
}

func (its DatatypeErrorCode) New(l *log.OrtooLog, args ...interface{}) OrtooError {
	format := fmt.Sprintf("[DatatypeError: %d] %s", its, datatypeErrFormats[its])
	err := &OrtooErrorImpl{
		Code: ErrorCode(its),
		Msg:  fmt.Sprintf(format, args...),
	}
	if l != nil {
		_ = l.OrtooSkipErrorf(err, 3, err.Msg)
	} else {
		_ = log.OrtooErrorWithSkip(err, 3, err.Msg)
	}
	return err
}

func (its DatatypeErrorCode) ToErrorCode() ErrorCode {
	return ErrorCode(its)
}
