package model

import (
	"fmt"
	"github.com/knowhunger/ortoo/commons/log"
)

var (
	ReplyMessageFormat = map[TypeReplyStates]string{
		TypeReplyStates_ERR_CLIENT_INVALID_COLLECTION: "invalid collection: in server %s but %s",
	}
)

func NewClientReply(seq uint32, state TypeReplyStates, args ...interface{}) *ClientReply {
	if state != TypeReplyStates_OK {
		log.Logger.Errorf(ReplyMessageFormat[state], args)
	}
	return &ClientReply{
		Header: &RequestHeader{
			Version: ProtocolVersion,
			Seq:     seq,
			Type:    &RequestHeader_TypeReply{TypeReplies_CLIENT_REPLY},
		},
		State: &ReplyState{
			State: state,
			Msg:   fmt.Sprintf(ReplyMessageFormat[state], args),
		},
	}
}
