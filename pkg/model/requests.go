package model

import (
	"fmt"
	"strings"
)

const (
	pushPullHeadFormat = "|(%d)[%s][%d]|"
	clientHeadFormat   = "|[%s]|"
)

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
func (its *ClientRequest) ToString() string {
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, clientHeadFormat, its.Header.ToString())
	b.WriteString(its.Client.ToString())
	return b.String()
}
