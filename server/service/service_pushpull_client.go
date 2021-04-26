package service

import (
	gocontext "context"
	"fmt"
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/server/constants"
	"github.com/knowhunger/ortoo/server/svrcontext"
	"reflect"
)

// ProcessPushPull processes a GRPC for Push-Pull
func (its *OrtooService) ProcessPushPull(goCtx gocontext.Context, in *model.PushPullMessage) (*model.PushPullMessage, error) {
	ctx := svrcontext.NewServerContext(goCtx, constants.TagPushPull).UpdateClient(in.Cuid)
	collectionDoc, rpcErr := its.getCollectionDocWithRPCError(ctx, in.Collection)
	if rpcErr != nil {
		return nil, rpcErr
	}
	ctx.UpdateCollection(collectionDoc.GetSummary())

	clientDoc, err := its.mongo.GetClient(ctx, in.Cuid)
	if err != nil {
		return nil, errors.NewRPCError(err)
	}
	if clientDoc == nil {
		msg := fmt.Sprintf("no client '%s:%s'", in.Collection, in.Cuid)
		return nil, errors.NewRPCError(errors.ServerNoResource.New(ctx.L(), msg))
	}
	ctx.UpdateClient(clientDoc.ToString())
	ctx.L().Infof("REQ[PUPU] %v", in.ToString())
	if clientDoc.CollectionNum != collectionDoc.Num {
		msg := fmt.Sprintf("client '%s' accesses collection(%d)", clientDoc.ToString(), collectionDoc.Num)
		return nil, errors.NewRPCError(errors.ServerNoPermission.New(ctx.L(), msg))
	}

	response := &model.PushPullMessage{
		Header:     in.Header,
		Collection: in.Collection,
		Cuid:       in.Cuid,
	}

	var chanList []<-chan *model.PushPullPack

	for _, ppp := range in.PushPullPacks {
		handler := newPushPullHandler(ctx, ppp, clientDoc, collectionDoc, its)
		chanList = append(chanList, handler.Start())
	}
	remainingChan := len(chanList)
	cases := make([]reflect.SelectCase, remainingChan)
	for i, ch := range chanList {
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
	}

	for remainingChan > 0 && chanList != nil {
		_, value, ok := reflect.Select(cases)
		remainingChan--
		if !ok {
			continue
		} else {
			ppp := value.Interface().(*model.PushPullPack)
			response.PushPullPacks = append(response.PushPullPacks, ppp)
		}
	}
	ctx.L().Infof("RES[PUPU] %v", response.ToString())
	return response, nil
}
