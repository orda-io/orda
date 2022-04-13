package errors

import "github.com/orda-io/orda/pkg/log"

// ErrorCode is a type for error code of OrdaError
type ErrorCode uint32

// New creates an error related to the datatype
func (its ErrorCode) New(l *log.OrdaLog, args ...interface{}) OrdaError {
	code := uint32(its) / 100
	switch code {
	case 1:
		return newSingleOrdaError(l, its, "ClientError", clientErrFormats[its], args...)
	case 2:
		return newSingleOrdaError(l, its, "DatatypeError", datatypeErrFormats[its], args...)
	case 3:
		return newSingleOrdaError(l, its, "PushPullError", pushPullErrFormats[its], args...)
	case 4:
		return newSingleOrdaError(l, its, "ServerError", serverErrFormats[its], args...)

	}
	panic("Unsupported error")
}

const (
	baseBasicCode    ErrorCode = 0
	baseClientCode   ErrorCode = 100
	baseDatatypeCode ErrorCode = 200
	basePushPullCode ErrorCode = 300
	baseServerCode   ErrorCode = 400
)

const (
	// MultipleErrors is an error code that includes many OrdaErrors
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
	DatatypeIllegalOperation
	DatatypeInvalidParent
	DatatypeNoOp
	DatatypeMarshal
	DatatypeNoTarget
	DatatypeInvalidPatch
)

var datatypeErrFormats = map[ErrorCode]string{
	DatatypeCreate:            "fail to create datatype: %s",
	DatatypeSubscribe:         "fail to subscribe datatype: %s",
	DatatypeTransaction:       "fail to proceed transaction: %v",
	DatatypeSnapshot:          "fail to make a snapshot: %s",
	DatatypeIllegalParameters: "fail to execute the operations due to illegal parameters: %v",
	DatatypeIllegalOperation:  "fail to execute the illegal operation for %v: %v",
	DatatypeInvalidParent:     "fail to modify due to the invalid parent: %v",
	DatatypeNoOp:              "fail to issue operation: %v",
	DatatypeMarshal:           "fail to (un)marshal: %v",
	DatatypeNoTarget:          "fail to find target: %v",
	DatatypeInvalidPatch:      "fail to patch: %v",
}

// ServerXXX denotes the error when Server is running.
const (
	ServerDBQuery = baseServerCode + iota
	ServerDBInit
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
	PushPullAbortionOfServer = basePushPullCode + iota
	PushPullAbortionOfClient
	PushPullDuplicateKey
	PushPullMissingOps
	PushPullNoDatatypeToSubscribe
)

var pushPullErrFormats = map[ErrorCode]string{
	PushPullAbortionOfServer:      "aborted push-pull due to server: %v",
	PushPullAbortionOfClient:      "aborted push-pull due to client: %v",
	PushPullDuplicateKey:          "duplicate datatype key: %v",
	PushPullMissingOps:            "aborted push due to missing operations: %v",
	PushPullNoDatatypeToSubscribe: "no datatype to subscribe: %v",
}
