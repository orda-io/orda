package model

func NewPushPullRequest(id int) *PushPullRequest {
	return &PushPullRequest{
		Header: &RequestHeader{
			Version: Version,
			Type:    TypeRequests_PUSHPULL_REQUEST,
		},
		Id:            0,
		PushPullPacks: nil,
	}
}
