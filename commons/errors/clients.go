package errors

import (
	"fmt"
	"github.com/knowhunger/ortoo/commons/log"
)

type errorCodeClient uint32

const baseClientCode errorCodeClient = 100
const (
	ErrClientNotConnected = baseClientCode + iota
	ErrClientConnect
)

var clientErrFormats = map[errorCodeClient]string{
	ErrClientNotConnected: "%s: client is not connected",
	ErrClientConnect:      "fail to connect: %s ",
}

type ClientError struct {
	code errorCodeClient
	msg  string
}

func (c *ClientError) Error() string {
	return c.msg
}

func NewClientError(code errorCodeClient, args ...interface{}) *ClientError {
	format := fmt.Sprintf("[ClientError: %d] %s", code, clientErrFormats[code])
	err := &ClientError{
		code: code,
		msg:  fmt.Sprintf(format, args...),
	}
	_ = log.OrtooErrorWithSkip(err, 3, err.msg)
	return err
}
