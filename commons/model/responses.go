package model

import (
	"fmt"
	"github.com/knowhunger/ortoo/commons/log"
)

var (
	replyMessageFormat = map[StateOfResponse]string{
		StateOfResponse_ERR_CLIENT_INVALID_COLLECTION: "invalid collection: in server %s but %s",
	}
)

//NewClientReply creates a new ClientReply
func NewClientReply(seq uint32, state StateOfResponse, args ...interface{}) *ClientResponse {
	if state != StateOfResponse_OK {
		log.Logger.Errorf(replyMessageFormat[state], args)
	}
	return &ClientResponse{
		Header: &RequestHeader{
			Version: ProtocolVersion,
			Seq:     seq,
			TypeOf:  &RequestHeader_TypeOfReply{TypeOfResponse_CLIENT_REPLY},
		},
		State: &ResponseState{
			State: state,
			Msg:   fmt.Sprintf(replyMessageFormat[state], args),
		},
	}
}
