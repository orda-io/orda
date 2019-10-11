package service

import (
	"context"
	"encoding/hex"
	"github.com/golang/protobuf/proto"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
	"github.com/knowhunger/ortoo/server/mongodb"
	"github.com/knowhunger/ortoo/server/mongodb/schema"
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
		handler := NewPushPullHandler(ctx, o.mongo, ppp, clientDoc)
		chanList = append(chanList, handler.Start())
	}
	remainingChan := len(chanList)
	cases := make([]reflect.SelectCase, remainingChan)
	for i, ch := range chanList {
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
	}
	for remainingChan > 0 {
		chosen, value, ok := reflect.Select(cases)
		if !ok {
			_ = log.OrtooErrorf(nil, "fail to run")
		}
		ch := chanList[chosen]
		msg := value.Interface()

		log.Logger.Infof("%v %v", ch, msg)
	}

	return &model.PushPullResponse{Id: in.Id}, nil
}

func NewPushPullHandler(ctx context.Context, mongo *mongodb.RepositoryMongo, ppp *model.PushPullPack, clientDoc *schema.ClientDoc) *PushPullHandler {
	return &PushPullHandler{
		ctx:          ctx,
		mongo:        mongo,
		pushPullPack: ppp,
	}
}

type PushPullHandler struct {
	ctx context.Context

	mongo        *mongodb.RepositoryMongo
	pushPullPack *model.PushPullPack
	duid         string
}

func (p *PushPullHandler) Start() chan *model.PushPullPack {
	retCh := make(chan *model.PushPullPack)
	go p._start(retCh)
	return retCh
}

func (p *PushPullHandler) _start(retCh <-chan *model.PushPullPack) {
	p.duid = hex.EncodeToString(p.pushPullPack.Duid)

	//p.mongo.GetDatatype()
}
