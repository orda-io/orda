package model

func NewPushPullRequest(seq uint32) *PushPullRequest {
	return &PushPullRequest{
		Header: &RequestHeader{
			Version: ProtocolVersion,
			Seq:     seq,
			Type:    &RequestHeader_TypeRequest{TypeRequests_PUSHPULL_REQUEST},
		},
		PushPullPacks: nil,
	}
}

func NewClientRequest(client *Client, seq uint32) *ClientRequest {
	return &ClientRequest{
		Header: &RequestHeader{
			Version: ProtocolVersion,
			Seq:     seq,
			Type:    &RequestHeader_TypeRequest{TypeRequests_CLIENTCREATE_REQUEST},
		},
		Client: client,
	}
}

func NewClientCreateReply(seq uint32) *ClientReply {
	return &ClientReply{
		Header: &RequestHeader{
			Version: ProtocolVersion,
			Seq:     seq,
			Type:    &RequestHeader_TypeReply{TypeReplies_CLIENTCREATE_REPLY},
		},
	}
}
