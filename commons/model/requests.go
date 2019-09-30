package model

func NewRequestHeader(seq uint32, typeOf TypeOfMessage, collection string) *MessageHeader {
	return &MessageHeader{
		Version:    ProtocolVersion,
		Seq:        seq,
		TypeOf:     typeOf,
		Collection: collection,
	}
}

//NewPushPullRequest creates a new PushPullRequest
func NewPushPullRequest(seq uint32, collection string, pushPullPackList ...*PushPullPack) *PushPullRequest {
	return &PushPullRequest{
		Header:        NewRequestHeader(seq, TypeOfMessage_REQUEST_PUSHPULL, collection),
		PushPullPacks: pushPullPackList,
	}
}

//NewClientRequest creates a new ClientRequest
func NewClientRequest(seq uint32, collection string, client *Client) *ClientRequest {
	return &ClientRequest{
		Header: NewRequestHeader(seq, TypeOfMessage_REQUEST_CLIENT, collection),
		Client: client,
	}
}
