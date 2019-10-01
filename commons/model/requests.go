package model

func NewMessageHeader(seq uint32, typeOf TypeOfMessage, collection string, cuid []byte) *MessageHeader {
	return &MessageHeader{
		Version:    ProtocolVersion,
		Seq:        seq,
		TypeOf:     typeOf,
		Collection: collection,
		Cuid:       cuid,
	}
}

//NewPushPullRequest creates a new PushPullRequest
func NewPushPullRequest(seq uint32, client *Client, pushPullPackList ...*PushPullPack) *PushPullRequest {
	return &PushPullRequest{
		Header:        NewMessageHeader(seq, TypeOfMessage_REQUEST_PUSHPULL, client.Collection, client.Cuid),
		PushPullPacks: pushPullPackList,
	}
}

//NewClientRequest creates a new ClientRequest
func NewClientRequest(seq uint32, client *Client) *ClientRequest {
	return &ClientRequest{
		Header: NewMessageHeader(seq, TypeOfMessage_REQUEST_CLIENT, client.Collection, client.Cuid),
		Client: client,
	}
}
