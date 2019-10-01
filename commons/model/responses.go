package model

import (
	"fmt"
	"github.com/knowhunger/ortoo/commons/log"
)

var (
	responseFormat = map[StateOfResponse]string{
		StateOfResponse_ERR_CLIENT_INVALID_COLLECTION: "invalid collection: in server %s but %s",
	}
)

//NewClientResponse creates a new ClientReply
func NewClientResponse(header *MessageHeader, state StateOfResponse, args ...interface{}) *ClientResponse {
	if state != StateOfResponse_OK {
		log.Logger.Errorf(responseFormat[state], args)
	}
	return &ClientResponse{
		Header: NewMessageHeader(header.Seq, TypeOfMessage_RESPONSE_CLIENT, header.Collection, header.Cuid),
		State: &ResponseState{
			State: state,
			Msg:   fmt.Sprintf(responseFormat[state], args),
		},
	}
}
