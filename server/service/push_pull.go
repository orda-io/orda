package service

import (
	"context"
	"github.com/golang/protobuf/proto"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
)

func (o *OrtooService) ProcessPushPull(ctx context.Context, in *model.PushPullRequest) (*model.PushPullReply, error) {
	log.Logger.Infof("Received: %v", proto.MarshalTextString(in))
	return &model.PushPullReply{Id: in.Id}, nil
}
