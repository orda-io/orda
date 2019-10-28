package model

import (
	"fmt"
	"github.com/knowhunger/ortoo/commons/log"
)

type rpcErrorCode uint32

const (
	//RPCErrMongoDB is the error related to MongoDB
	RPCErrMongoDB rpcErrorCode = iota
	//RPCErrClientInconsistentCollection is the error when a client has different collection with the previously
	RPCErrClientInconsistentCollection
	//RPCErrNoClient is the error when the given client does not exist.
	RPCErrNoClient
)

var formatMap = map[rpcErrorCode]string{
	RPCErrMongoDB:                      "MongoDB is not working",
	RPCErrClientInconsistentCollection: "invalid collections: %s (server) vs. %s (client)",
	RPCErrNoClient:                     "No client exists in the server",
}

//RPCError defines errors of RPC
type RPCError struct {
	code rpcErrorCode
	msg  string
}

func newRPCError(code rpcErrorCode) *RPCError {
	return &RPCError{code: code}
}

func (r *RPCError) Error() string {
	return r.msg
}

//NewRPCError creates a new RPCError
func NewRPCError(code rpcErrorCode, args ...interface{}) *RPCError {
	format := fmt.Sprintf("[RPCErr: %d] ", code) + formatMap[code]
	err := &RPCError{
		code: code,
		msg:  fmt.Sprintf(format, args...),
	}
	_ = log.OrtooErrorWithSkip(err, 3, err.msg)
	return err
}
