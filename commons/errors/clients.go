package errors

import (
	"fmt"
	"github.com/knowhunger/ortoo/commons/log"
)

type errorCodeClient uint32

const baseClientCode errorCodeClient = 100

// ErrClientXXXX defines the error related to client
const (
	ErrClientNotConnected = baseClientCode + iota
	ErrClientConnect
	ErrClientClose
)

var clientErrFormats = map[errorCodeClient]string{
	ErrClientNotConnected: "%s: client is not connected",
	ErrClientConnect:      "fail to connect: %s ",
	ErrClientClose:        "fail to close: %s",
}

// ClientError is an error for Client
type ClientError struct {
	code errorCodeClient
	msg  string
}

func (c *ClientError) Error() string {
	return c.msg
}

// NewClientError creates an error related to the client
func NewClientError(code errorCodeClient, args ...interface{}) *ClientError {
	format := fmt.Sprintf("[ClientError: %d] %s", code, clientErrFormats[code])
	err := &ClientError{
		code: code,
		msg:  fmt.Sprintf(format, args...),
	}
	_ = log.OrtooErrorWithSkip(err, 3, err.msg)
	return err
}
