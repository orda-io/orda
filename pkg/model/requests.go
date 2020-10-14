package model

import (
	"fmt"
	"github.com/knowhunger/ortoo/pkg/types"
	"github.com/knowhunger/ortoo/pkg/utils"
	"strings"
)

const (
	pushPullHeadFormat = "|(%d)[%s][%d]|"
	clientHeadFormat   = "|[%s]|"
)

// NewMessageHeader generates a message header.
func NewMessageHeader(seq uint32, typeOf TypeOfMessage, collection string, clientAlias string, cuid []byte) *MessageHeader {
	return &MessageHeader{
		Version:     ProtocolVersion,
		Seq:         seq,
		TypeOf:      typeOf,
		Collection:  collection,
		ClientAlias: clientAlias,
		Cuid:        cuid,
	}
}

func (its *MessageHeader) GetClient() string {
	return fmt.Sprintf("%s(%s)", its.ClientAlias, types.UIDtoString(its.Cuid))
}

func (its *MessageHeader) GetClientSummary() string {
	return utils.MakeSummary(its.ClientAlias, its.Cuid, "C")
}

// NewPushPullRequest creates a new PushPullRequest
func NewPushPullRequest(seq uint32, client *Client, pushPullPackList ...*PushPullPack) *PushPullRequest {
	return &PushPullRequest{
		Header:        NewMessageHeader(seq, TypeOfMessage_REQUEST_PUSHPULL, client.Collection, client.Alias, client.CUID),
		PushPullPacks: pushPullPackList,
	}
}

// ToString returns customized string
func (its *PushPullRequest) ToString() string {
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, pushPullHeadFormat, its.ID, its.Header.ToString(), len(its.PushPullPacks))
	for _, ppp := range its.PushPullPacks {
		b.WriteString(" ")
		b.WriteString(ppp.ToString())
	}
	return b.String()
}

// NewClientRequest creates a new ClientRequest
func NewClientRequest(seq uint32, client *Client) *ClientRequest {
	return &ClientRequest{
		Header: NewMessageHeader(seq, TypeOfMessage_REQUEST_CLIENT, client.Collection, client.Alias, client.CUID),
		Client: client,
	}
}

// ToString returns customized string
func (c *ClientRequest) ToString() string {
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, clientHeadFormat, c.Header.ToString())
	b.WriteString(c.Client.ToString())
	return b.String()
}
