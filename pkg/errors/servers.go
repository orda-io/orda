package errors

import (
	"fmt"
	"github.com/knowhunger/ortoo/pkg/log"
	"github.com/ztrue/tracerr"
)

// NewPushPullError generates a PushPullError.
func (its ServerErrorCode) New(l *log.OrtooLog, args ...interface{}) OrtooError {
	format := fmt.Sprintf("[PushPullError: %d] %s", its, serverErrFormats[its])
	err := &singleOrtooError{
		tError: tracerr.New(fmt.Sprintf(format, args...)),
		Code:   ErrorCode(its),
	}
	err.Print(l)
	return err
}

func (its ServerErrorCode) ec() ErrorCode {
	return ErrorCode(its)
}
