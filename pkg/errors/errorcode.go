package errors

import "github.com/knowhunger/ortoo/pkg/log"

// ErrorCode is a type for error code of OrtooError
type ErrorCode uint32

// New creates an error related to the datatype
func (its ErrorCode) New(l *log.OrtooLog, args ...interface{}) OrtooError {
	code := uint32(its) / 100
	switch code {
	case 1:
		return newSingleOrtooError(l, its, "ClientError", clientErrFormats[its], args...)
	case 2:
		return newSingleOrtooError(l, its, "DatatypeError", datatypeErrFormats[its], args...)
	case 3:
		return newSingleOrtooError(l, its, "ServerError", serverErrFormats[its], args)
	case 4:
		return newSingleOrtooError(l, its, "PushPullError", pushPullErrFormats[its], args)
	}
	panic("Unsupported error")
}

const (
	baseBasicCode    ErrorCode = 0
	baseClientCode   ErrorCode = 100
	baseDatatypeCode ErrorCode = 200
	baseServerCode   ErrorCode = 300
	basePushPullCode ErrorCode = 400
)

const (
	// MultipleErrors is an error code that includes many OrtooErrors
	MultipleErrors = baseBasicCode + iota
)

// ErrClientXXXX defines the error related to client
const (
	ClientConnect = baseClientCode + iota
	ClientClose
	ClientSync
)

var clientErrFormats = map[ErrorCode]string{
	ClientConnect: "fail to connect: %v",
	ClientClose:   "fail to close: %v",
	ClientSync:    "fail to sync: %v",
}

// DatatypeXXX defines an error related to Datatype
const (
	DatatypeCreate = baseDatatypeCode + iota
	DatatypeSubscribe
	DatatypeTransaction
	DatatypeSnapshot
	DatatypeIllegalParameters
	DatatypeInvalidParent
	DatatypeNoOp
	DatatypeMarshal
	DatatypeNoTarget
)

var datatypeErrFormats = map[ErrorCode]string{
	DatatypeCreate:            "fail to create datatype: %s",
	DatatypeSubscribe:         "fail to subscribe datatype: %s",
	DatatypeTransaction:       "fail to proceed transaction: %v",
	DatatypeSnapshot:          "fail to make a snapshot: %s",
	DatatypeIllegalParameters: "fail to execute the operation due to illegal operation: %v",
	DatatypeInvalidParent:     "fail to modify due to the invalid parent: %v",
	DatatypeNoOp:              "fail to issue operation: %v",
	DatatypeMarshal:           "fail to (un)marshal: %v",
	DatatypeNoTarget:          "fail to find target: %v",
}

// ServerXXX denotes the error when Server is running.
const (
	ServerDBQuery = baseServerCode + iota
	ServerDBDecode
	ServerNoResource
	ServerNoPermission
	ServerInit
	ServerNotify
)

var serverErrFormats = map[ErrorCode]string{
	ServerDBQuery:      "fail to succeed DB query: %v",
	ServerDBDecode:     "fail to decode in DB: %v",
	ServerNoResource:   "find no resource: %v",
	ServerNoPermission: "have no permission: %v",
	ServerInit:         "fail to initialize server: %v",
	ServerNotify:       "fail to notify push-pull: %v",
}

const (
	PushPullAbortedForServer = basePushPullCode + iota
	PushPullAbortedForClient
	PushPullDuplicateKey
	PushPullMissingOps
	PushPullNoDatatypeToSubscribe
)

var pushPullErrFormats = map[ErrorCode]string{
	PushPullAbortedForServer:      "aborted push-pull due to server: %v",
	PushPullAbortedForClient:      "aborted push-pull due to client: %v",
	PushPullDuplicateKey:          "duplicate datatype key: %v",
	PushPullMissingOps:            "aborted push due to missing operations: %v",
	PushPullNoDatatypeToSubscribe: "no datatype to subscribe: %v",
}
