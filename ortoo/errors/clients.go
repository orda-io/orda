package errors

import (
	"fmt"
	"github.com/knowhunger/ortoo/ortoo/log"
)

type ClientErrorCode ErrorCode

const baseClientCode ClientErrorCode = 100

// ErrClientXXXX defines the error related to client
const (
	ErrClientNotConnected = baseClientCode + iota
	ErrClientConnect
	ErrClientClose
)

var clientErrFormats = map[ClientErrorCode]string{
	ErrClientNotConnected: "%s: client is not connected",
	ErrClientConnect:      "fail to connect: %s ",
	ErrClientClose:        "fail to close: %s",
}

// NewClientError creates an error related to the client
func (its ClientErrorCode) New(args ...interface{}) OrtooError {
	format := fmt.Sprintf("[ClientError: %d] %s", its, clientErrFormats[its])
	err := &OrtooErrorImpl{
		Code: ErrorCode(its),
		Msg:  fmt.Sprintf(format, args...),
	}
	_ = log.OrtooErrorWithSkip(err, 3, err.Msg)
	return err
}
