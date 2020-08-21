package errors

import (
	"fmt"
	"github.com/knowhunger/ortoo/ortoo/log"
)

type ErrorCode uint32

const baseDatatypeCode ErrorCode = 200

func (its ErrorCode) New(args ...interface{}) *OrtooError {
	err := &OrtooError{
		Code: its,
		Msg:  fmt.Sprintf(datatypeErrFormats[its], args...),
	}
	_ = log.OrtooErrorWithSkip(err, 3, err.Msg)
	return err
}

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

var datatypeErrFormats = map[ErrorCode]string{
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

// OrtooError is an error related to Datatype
type OrtooError struct {
	Code ErrorCode
	Msg  string
}

func ToOrtooError(err error) *OrtooError {
	if dErr, ok := err.(*OrtooError); ok {
		return dErr
	}
	return nil
}

func (d *OrtooError) Error() string {
	return d.Msg
}

// New creates an error related to the datatype
func New(code ErrorCode, args ...interface{}) *OrtooError {
	format := fmt.Sprintf("[OrtooError: %d] %s", code, datatypeErrFormats[code])
	err := &OrtooError{
		Code: code,
		Msg:  fmt.Sprintf(format, args...),
	}
	_ = log.OrtooErrorWithSkip(err, 3, err.Msg)
	return err
}
