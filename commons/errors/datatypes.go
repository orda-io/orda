package errors

import (
	"fmt"
	"github.com/knowhunger/ortoo/commons/log"
)

type errorCodeDatatype uint32

const baseDatatypeCode errorCodeDatatype = 200
const (
	ErrDatatypeCreate = baseDatatypeCode + iota
	ErrDatatypeSubscribe
	ErrDatatypeTransaction
)

var datatypeErrFormats = map[errorCodeDatatype]string{
	ErrDatatypeCreate:      "fail to create datatype: %s",
	ErrDatatypeSubscribe:   "fail to subscribe datatype: %s",
	ErrDatatypeTransaction: "fail to proceed transaction: %s",
}

type DatatypeError struct {
	code errorCodeDatatype
	msg  string
}

func (d *DatatypeError) Error() string {
	return d.msg
}

func NewDatatypeError(code errorCodeDatatype, args ...interface{}) *DatatypeError {
	format := fmt.Sprintf("[DatatypeError: %d] %s", code, datatypeErrFormats[code])
	err := &DatatypeError{
		code: code,
		msg:  fmt.Sprintf(format, args...),
	}
	_ = log.OrtooErrorWithSkip(err, 3, err.msg)
	return err
}
