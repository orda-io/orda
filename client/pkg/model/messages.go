package model

import (
	"fmt"
	"strings"
)

const (
	clientHeadFormat = "[%s|%s|%s]"
)

// NewPushPullMessage creates a new PushPullRequest
func NewPushPullMessage(seq uint32, client *Client, pushPullPackList ...*PushPullPack) *PushPullMessage {
	return &PushPullMessage{
		Header:        NewMessageHeader(RequestType_PUSHPULLS),
		Collection:    client.Collection,
		Cuid:          client.CUID,
		PushPullPacks: pushPullPackList,
	}
}

// ToString returns customized string
func (its *PushPullMessage) ToString(isFull bool) string {
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, "Head[%s] PushPullPack[%d]{", its.Header.ToString(), len(its.PushPullPacks))
	for _, ppp := range its.PushPullPacks {
		b.WriteString(" ")
		b.WriteString(ppp.ToString(isFull))
	}
	b.WriteString("}")
	return b.String()
}

// GetClient returns the model of the client
func (its *PushPullMessage) GetClient() *Client {
	return &Client{
		CUID:       its.Cuid,
		Alias:      "",
		Collection: its.Collection,
		SyncType:   SyncType_LOCAL_ONLY,
	}
}

// NewClientMessage creates a new ClientRequest
func NewClientMessage(client *Client) *ClientMessage {
	return &ClientMessage{
		Header:      NewMessageHeader(RequestType_CLIENTS),
		Collection:  client.Collection,
		Cuid:        client.CUID,
		ClientAlias: client.Alias,
		SyncType:    client.SyncType,
	}
}

// ToString returns customized string
func (its *ClientMessage) ToString() string {
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, clientHeadFormat, its.Header.ToString(), its.Collection, its.Cuid)
	b.WriteString(" SyncType:")
	b.WriteString(its.SyncType.String())
	return b.String()
}

// GetClient returns the model of client
func (its *ClientMessage) GetClient() *Client {
	return &Client{
		CUID:       its.Cuid,
		Alias:      its.ClientAlias,
		Collection: its.Collection,
		Type:       its.ClientType,
		SyncType:   its.SyncType,
	}
}

// GetClientSummary returns the summary of client
func (its *ClientMessage) GetClientSummary() string {
	return fmt.Sprintf("%s(%s)", its.ClientAlias, its.Cuid)
}
