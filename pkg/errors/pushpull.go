package errors

import (
	"fmt"
	"github.com/knowhunger/ortoo/pkg/log"
)

// PushPullError defines the errors used in datatype.
type PushPullError struct {
	Code PushPullErrorCode
	Msg  string
}

func (its *PushPullError) Error() string {
	return fmt.Sprintf("[PushPullErr: %d] %s", its.Code, its.Msg)
}

func (its PushPullErrorCode) New(l *log.OrtooLog, args ...interface{}) OrtooError {
	return newSingleOrtooError(l, ErrorCode(its), "PushPullError", pushPullErrFormats[its], args...)
}
