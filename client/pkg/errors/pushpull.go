package errors

import (
	"fmt"
)

// PushPullError defines the errors used in datatype.
type PushPullError struct {
	Code ErrorCode
	Msg  string
}

func (its *PushPullError) Error() string {
	return fmt.Sprintf("[PushPullErr: %d] %s", its.Code, its.Msg)
}
