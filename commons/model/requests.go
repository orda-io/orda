package model

import (
	"fmt"
	"strings"
)

const (
	pushPullHeadFormat = "|HEAD:[%s] ID=%d LEN=%d|"
	clientHeadFormat   = "|HEAD:[%s]|"
)

func NewMessageHeader(seq uint32, typeOf TypeOfMessage, collection string, cuid []byte) *MessageHeader {
	return &MessageHeader{
		Version:    ProtocolVersion,
		Seq:        seq,
		TypeOf:     typeOf,
		Collection: collection,
		Cuid:       cuid,
	}
}

// NewPushPullRequest creates a new PushPullRequest
func NewPushPullRequest(seq uint32, client *Client, pushPullPackList ...*PushPullPack) *PushPullRequest {
	return &PushPullRequest{
		Header:        NewMessageHeader(seq, TypeOfMessage_REQUEST_PUSHPULL, client.Collection, client.CUID),
		PushPullPacks: pushPullPackList,
	}
}

func (p *PushPullRequest) ToString() string {
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, pushPullHeadFormat, p.Header.ToString(), p.ID, len(p.PushPullPacks))
	for _, ppp := range p.PushPullPacks {
		b.WriteString(" ")
		b.WriteString(ppp.ToString())
	}
	return b.String()
}

// NewClientRequest creates a new ClientRequest
func NewClientRequest(seq uint32, client *Client) *ClientRequest {
	return &ClientRequest{
		Header: NewMessageHeader(seq, TypeOfMessage_REQUEST_CLIENT, client.Collection, client.CUID),
		Client: client,
	}
}

func (c *ClientRequest) ToString() string {
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, clientHeadFormat, c.Header.ToString())
	b.WriteString(c.Client.ToString())
	return b.String()
}
