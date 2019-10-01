package service

import (
	"context"
	"encoding/hex"
	"github.com/golang/protobuf/proto"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
	"reflect"
)

//ProcessPushPull processes a GRPC for Push-Pull
func (o *OrtooService) ProcessPushPull(ctx context.Context, in *model.PushPullRequest) (*model.PushPullResponse, error) {
	log.Logger.Infof("Received: %v, %s", proto.MarshalTextString(in), hex.EncodeToString(in.Header.GetCuid()))
	collectionDoc, err := o.mongo.GetCollection(ctx, in.Header.GetCollection())
	if collectionDoc == nil || err != nil {
		return nil, model.NewRPCError(model.RPCErrMongoDB)
	}

	clientDoc, err := o.mongo.GetClient(ctx, hex.EncodeToString(in.Header.GetCuid()))
	if err != nil {
		return nil, model.NewRPCError(model.RPCErrMongoDB)
	}
	if clientDoc == nil {
		return nil, model.NewRPCError(model.RPCErrNoClient)
	}
	if clientDoc.Collection != collectionDoc.Name {
		return nil, model.NewRPCError(model.RPCErrClientInconsistentCollection, clientDoc.Collection, collectionDoc.Name)
	}
	var chanList []chan *model.PushPullPack
	for _, ppp := range in.PushPullPacks {
		handler := NewPushPullHandler(ppp)
		chanList = append(chanList, handler.Start())
	}
	cases := make([]reflect.SelectCase, len(chanList))
	for i, ch := range chanList {
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
	}
	//chosen, value, ok := reflect.Select(cases)
	//ch := chans[chosen]
	//msg := value.String()

	return &model.PushPullResponse{Id: in.Id}, nil
}

func NewPushPullHandler(ppp *model.PushPullPack) *PushPullHandler {
	return &PushPullHandler{pushPullPack: ppp}
}

type PushPullHandler struct {
	pushPullPack *model.PushPullPack
}

func (p *PushPullHandler) Start() chan *model.PushPullPack {
	return nil
}
