package errors

import (
	"github.com/knowhunger/ortoo/pkg/log"
)

// New creates an error related to the client
func (its ClientErrorCode) New(l *log.OrtooLog, args ...interface{}) OrtooError {
	return newSingleOrtooError(l, ErrorCode(its), "ClientError", clientErrFormats[its], args...)
}
