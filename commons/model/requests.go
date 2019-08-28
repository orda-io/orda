package model

func NewPushPullRequest(id int32) *PushPullRequest {
	return &PushPullRequest{
		Header: &RequestHeader{
			Version: ProtocolVersion,
			Type:    &RequestHeader_TypeRequest{TypeRequests_PUSHPULL_REQUEST},
		},
		Id:            id,
		PushPullPacks: nil,
	}
}

func NewClientCreateRequest(client *Client) *ClientCreateRequest {
	return &ClientCreateRequest{
		Header: &RequestHeader{
			Version: ProtocolVersion,
			Type:    &RequestHeader_TypeRequest{TypeRequests_CLIENTCREATE_REQUEST},
		},
		Client: client,
	}
}

func NewClientCreateReply() *ClientCreateReply {
	return &ClientCreateReply{
		Header: &RequestHeader{
			Version: ProtocolVersion,
			Type:    &RequestHeader_TypeReply{TypeReplies_CLIENTCREATE_REPLY},
		},
	}
}
