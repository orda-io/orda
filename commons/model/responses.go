package model

import (
	"fmt"
	"github.com/knowhunger/ortoo/commons/log"
)

var (
	replyMessageFormat = map[TypeResponseStates]string{
		TypeResponseStates_ERR_CLIENT_INVALID_COLLECTION: "invalid collection: in server %s but %s",
	}
)

//NewClientReply creates a new ClientReply
func NewClientReply(seq uint32, state TypeResponseStates, args ...interface{}) *ClientResponse {
	if state != TypeResponseStates_OK {
		log.Logger.Errorf(replyMessageFormat[state], args)
	}
	return &ClientResponse{
		Header: &RequestHeader{
			Version: ProtocolVersion,
			Seq:     seq,
			Type:    &RequestHeader_TypeReply{TypeResponse_CLIENT_REPLY},
		},
		State: &ResponseState{
			State: state,
			Msg:   fmt.Sprintf(replyMessageFormat[state], args),
		},
	}
}
