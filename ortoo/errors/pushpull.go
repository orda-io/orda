package errors

import (
	"fmt"
	"github.com/knowhunger/ortoo/ortoo/log"
)

// ErrorCodePushPull defines error codes for PushPull
type ErrorCodePushPull uint32

// PushPullErrXXX denotes the error when PushPull is processed.
const (
	PushPullErrQueryToDB ErrorCodePushPull = iota
	PushPullErrIllegalFormat
	PushPullErrDuplicateDatatypeKey
	PushPullErrPullOperations
	PushPullErrPushOperations
	PushPullErrMissingOperations
	PushPullErrUpdateSnapshot
)

var pushPullMap = map[ErrorCodePushPull]string{
	PushPullErrQueryToDB:            "fail to query to DB: %v",
	PushPullErrIllegalFormat:        "illegal format: %s - %s",
	PushPullErrDuplicateDatatypeKey: "duplicate datatype key",
	PushPullErrPullOperations:       "fail to pull operations: %v",
	PushPullErrPushOperations:       "fail to push operations: %v",
	PushPullErrMissingOperations:    "fail to push due to missing operations: %v",
	PushPullErrUpdateSnapshot:       "fail to update snapshot: %v",
}

// PushPullError defines PushPullError.
type PushPullError struct {
	Code ErrorCodePushPull
	Msg  string
}

func (p *PushPullError) Error() string {
	return fmt.Sprintf("PushPullErr: %d", p.Code)
}

// PushPullTag defines a PushPullTag.
type PushPullTag struct {
	CollectionName string
	Key            string
	DUID           string
}

// NewPushPullError generates a PushPullError.
func NewPushPullError(code ErrorCodePushPull, tag PushPullTag, args ...interface{}) *PushPullError {
	format := fmt.Sprintf("[%s][%s][%s] ", tag.CollectionName, tag.Key, tag.DUID) + pushPullMap[code]
	err := &PushPullError{
		Code: code,
		Msg:  fmt.Sprintf(format, args...),
	}
	_ = log.OrtooErrorWithSkip(err, 3, err.Msg)
	return err
}
