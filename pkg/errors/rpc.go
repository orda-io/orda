package errors

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

// NewRPCError creates a RPC error
func NewRPCError(oErr OrtooError) error {
	var c codes.Code
	code := oErr.GetCode()
	switch code {
	case ServerDBQuery.ec():
		c = codes.Unavailable // temporally unavailable
	case ServerDBDecode.ec():
		c = codes.Internal // something is broken
	case ServerNoResource.ec():
		c = codes.NotFound
	case ServerNoPermission.ec():
		c = codes.Unauthenticated
	}
	return status.Error(c, oErr.Error())
}
