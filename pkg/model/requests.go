package model

import (
	"fmt"
	"strings"
)

const (
	pushPullHeadFormat = "|[%s][%d]|"
	clientHeadFormat   = "|[%s]|"
)

// NewPushPullRequest creates a new PushPullRequest
func NewPushPullRequest(seq uint32, client *Client, pushPullPackList ...*PushPullPack) *PushPullRequest {
	return &PushPullRequest{
		Header:        NewMessageHeader(RequestType_PUSHPULLS, client.Collection, client.Alias, client.CUID),
		PushPullPacks: pushPullPackList,
	}
}

// ToString returns customized string
func (its *PushPullRequest) ToString() string {
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, pushPullHeadFormat, its.Header.ToString(), len(its.PushPullPacks))
	for _, ppp := range its.PushPullPacks {
		b.WriteString(" ")
		b.WriteString(ppp.ToString())
	}
	return b.String()
}

// NewClientRequest creates a new ClientRequest
func NewClientRequest(seq uint32, client *Client) *ClientRequest {
	return &ClientRequest{
		Header:   NewMessageHeader(RequestType_PUSHPULLS, client.Collection, client.Alias, client.CUID),
		SyncType: client.SyncType,
	}
}

// ToString returns customized string
func (its *ClientRequest) ToString() string {
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, clientHeadFormat, its.Header.ToString())
	b.WriteString(its.Header.ToString())
	b.WriteString(" SyncType:")
	b.WriteString(its.SyncType.String())
	return b.String()
}

func (its *ClientRequest) GetClient() *Client {
	return &Client{
		CUID:       its.Header.Cuid,
		Alias:      its.Header.ClientAlias,
		Collection: its.Header.Collection,
		SyncType:   its.SyncType,
	}
}
