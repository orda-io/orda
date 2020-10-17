package errors

// ErrorCode is a type for error code of OrtooError
type ErrorCode uint32

// ClientErrorCode is a type for client errors
type ClientErrorCode ErrorCode

// DatatypeErrorCode is a type for datatype errors
type DatatypeErrorCode ErrorCode

// ServerErrorCode defines error codes for Server
type ServerErrorCode ErrorCode

// PushPullErrorCode defines error codes for Protocol
type PushPullErrorCode ErrorCode

const (
	baseBasicCode    ErrorCode         = 0
	baseClientCode   ClientErrorCode   = 100
	baseDatatypeCode DatatypeErrorCode = 200
	baseServerCode   ServerErrorCode   = 300
	basePushPullCode PushPullErrorCode = 990
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

var clientErrFormats = map[ClientErrorCode]string{
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

var datatypeErrFormats = map[DatatypeErrorCode]string{
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

var serverErrFormats = map[ServerErrorCode]string{
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
)

var pushPullErrFormats = map[PushPullErrorCode]string{
	PushPullAbortedForServer: "aborted push-pull due to server: %v",
	PushPullAbortedForClient: "aborted push-pull due to client: %v",
	PushPullDuplicateKey:     "duplicate datatype key: %v",
	PushPullMissingOps:       "aborted push due to missing operations: %v",
}
