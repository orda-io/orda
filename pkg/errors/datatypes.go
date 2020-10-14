package errors

import (
	"github.com/knowhunger/ortoo/pkg/log"
)

// New creates an error related to the datatype
func (its DatatypeErrorCode) New(l *log.OrtooLog, args ...interface{}) OrtooError {
	return newSingleOrtooError(l, ErrorCode(its), "DatatypeError", datatypeErrFormats[its], args...)
}

// ToErrorCode casts DatatypeErrorCode to ErrorCode
func (its DatatypeErrorCode) ToErrorCode() ErrorCode {
	return ErrorCode(its)
}
