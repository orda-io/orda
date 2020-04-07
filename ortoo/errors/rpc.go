package errors

import (
	"fmt"
	"github.com/knowhunger/ortoo/ortoo/log"
)

type errorCodeRPC uint32

const (
	// RPCErrMongoDB is the error related to MongoDB
	RPCErrMongoDB errorCodeRPC = iota
	// RPCErrClientInconsistentCollection is the error when a client has different collection with the previously
	RPCErrClientInconsistentCollection
	// RPCErrNoClient is the error when the specified client does not exist.
	RPCErrNoClient
)

var formatMap = map[errorCodeRPC]string{
	RPCErrMongoDB:                      "work no MongoDB",
	RPCErrClientInconsistentCollection: "invalid collections: %s (server) vs. %s (client)",
	RPCErrNoClient:                     "exist no client in the server",
}

// RPCError defines errors of RPC
type RPCError struct {
	code errorCodeRPC
	msg  string
}

func (r *RPCError) Error() string {
	return r.msg
}

// NewRPCError creates a new RPCError.
func NewRPCError(code errorCodeRPC, args ...interface{}) *RPCError {
	format := fmt.Sprintf("[RPCErr: %d] ", code) + formatMap[code]
	err := &RPCError{
		code: code,
		msg:  fmt.Sprintf(format, args...),
	}
	_ = log.OrtooErrorWithSkip(err, 3, err.msg)
	return err
}
