package model

import (
	"fmt"
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
		msg = fmt.Sprintf(responseFormat[state], args)
	}
	return &ClientResponse{
		Header: NewMessageHeader(header.Seq, TypeOfMessage_RESPONSE_CLIENT, header.Collection, header.ClientAlias, header.Cuid),
		State: &ResponseState{
			State: state,
			Msg:   msg,
		},
	}
}

// ToString returns customized string
func (c *ClientResponse) ToString() string {
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, clientHeadFormat, c.Header.ToString())
	b.WriteString(c.State.State.String())
	b.WriteString(":")
	b.WriteString(c.State.Msg)
	return b.String()
}

// ToString returns customized string
func (p *PushPullResponse) ToString() string {
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, pushPullHeadFormat, p.ID, p.Header.ToString(), len(p.PushPullPacks))
	for _, ppp := range p.PushPullPacks {
		b.WriteString(" ")
		b.WriteString(ppp.ToString())
	}
	return b.String()
}
