package model

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

type errorCodePushPull uint32

// PushPullErrXXX denotes the error when PushPull is processed.
const (
	PushPullErrQueryToDB errorCodePushPull = iota
	PushPullErrIllegalFormat
	PushPullErrDuplicateDatatypeKey
	PushPullErrPullOperations
	PushPullErrPushOperations
	PushPullErrMissingOperations
)

var pushPullMap = map[errorCodePushPull]string{
	PushPullErrQueryToDB:            "fail to query to DB: %v",
	PushPullErrIllegalFormat:        "illegal format: %s - %s",
	PushPullErrDuplicateDatatypeKey: "duplicate datatype key",
	PushPullErrPullOperations:       "fail to pull operations: %v",
	PushPullErrPushOperations:       "fail to push operations: %v",
	PushPullErrMissingOperations:    "fail to push due to missing operations: %v",
}

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

// PushPullError defines PushPullError.
type PushPullError struct {
	Code errorCodePushPull
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
func NewPushPullError(code errorCodePushPull, tag PushPullTag, args ...interface{}) *PushPullError {
	format := fmt.Sprintf("[%s][%s][%s] ", tag.CollectionName, tag.Key, tag.DUID) + pushPullMap[code]
	err := &PushPullError{
		Code: code,
		Msg:  fmt.Sprintf(format, args...),
	}
	_ = log.OrtooErrorWithSkip(err, 3, err.Msg)
	return err
}
