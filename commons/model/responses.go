package model

import (
	"fmt"
	"github.com/knowhunger/ortoo/commons/log"
	"strings"
)

var (
	responseFormat = map[StateOfResponse]string{
		StateOfResponse_ERR_CLIENT_INVALID_COLLECTION: "invalid collection: in server %s but %s",
	}
)

// NewClientResponse creates a new ClientReply
func NewClientResponse(header *MessageHeader, state StateOfResponse, args ...interface{}) *ClientResponse {
	msg := ""
	if state != StateOfResponse_OK {
		log.Logger.Errorf(responseFormat[state], args)
		msg = fmt.Sprintf(responseFormat[state], args)
	}
	return &ClientResponse{
		Header: NewMessageHeader(header.Seq, TypeOfMessage_RESPONSE_CLIENT, header.Collection, header.Cuid),
		State: &ResponseState{
			State: state,
			Msg:   msg,
		},
	}
}

func (c *ClientResponse) ToString() string {
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, clientHeadFormat, c.Header.ToString())
	b.WriteString(c.State.State.String())
	b.WriteString(":")
	b.WriteString(c.State.Msg)
	return b.String()
}

func (p *PushPullResponse) ToString() string {
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, pushPullHeadFormat, p.Header.ToString(), p.ID, len(p.PushPullPacks))
	for _, ppp := range p.PushPullPacks {
		b.WriteString(" ")
		b.WriteString(ppp.ToString())
	}
	return b.String()
}
