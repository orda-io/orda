package model

func NewPushPullRequest(id int32) *PushPullRequest {
	return &PushPullRequest{
		Header: &RequestHeader{
			Version: Version,
			Type:    TypeRequests_PUSHPULL_REQUEST,
		},
		Id:            id,
		PushPullPacks: nil,
	}
}
