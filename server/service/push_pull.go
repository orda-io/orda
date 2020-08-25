package service

import (
	"context"
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/log"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/pkg/types"
	"reflect"
)

// ProcessPushPull processes a GRPC for Push-Pull
func (o *OrtooService) ProcessPushPull(ctx context.Context, in *model.PushPullRequest) (*model.PushPullResponse, error) {
	log.Logger.Infof("receive %v", in.ToString())
	collectionDoc, err := o.mongo.GetCollection(ctx, in.Header.GetCollection())
	if collectionDoc == nil || err != nil {
		return nil, errors.NewRPCError(errors.RPCErrMongoDB)
	}

	clientDoc, err := o.mongo.GetClient(ctx, types.ToUID(in.Header.GetCuid()))
	if err != nil {
		return nil, errors.NewRPCError(errors.RPCErrMongoDB)
	}
	if clientDoc == nil {
		return nil, errors.NewRPCError(errors.RPCErrNoClient)
	}
	if clientDoc.CollectionNum != collectionDoc.Num {
		return nil, errors.NewRPCError(errors.RPCErrClientInconsistentCollection, clientDoc.CollectionNum, collectionDoc.Name)
	}
	var chanList []<-chan *model.PushPullPack
	for _, ppp := range in.PushPullPacks {
		handler := &PushPullHandler{
			Key:             ppp.Key,
			DUID:            types.ToUID(ppp.DUID),
			CUID:            clientDoc.CUID,
			ctx:             ctx,
			mongo:           o.mongo,
			notifier:        o.notifier,
			clientDoc:       clientDoc,
			collectionDoc:   collectionDoc,
			gotPushPullPack: ppp,
			gotOption:       ppp.GetPushPullPackOption(),
		}
		chanList = append(chanList, handler.Start())
	}
	remainingChan := len(chanList)
	cases := make([]reflect.SelectCase, remainingChan)
	for i, ch := range chanList {
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
	}
	response := &model.PushPullResponse{
		Header: in.Header,
		ID:     in.ID,
	}

	for remainingChan > 0 && chanList != nil {
		_, value, ok := reflect.Select(cases)
		remainingChan--
		if !ok {
			_ = log.OrtooErrorf(nil, "fail to run")
			continue
		} else {
			// ch := chanList[chosen]
			msg := value.Interface()
			// log.Logger.Infof("%v %v", ch, msg)
			ppp := msg.(*model.PushPullPack)
			response.PushPullPacks = append(response.PushPullPacks, ppp)
		}
	}

	return response, nil
}
