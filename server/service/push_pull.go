package service

import (
	gocontext "context"
	"fmt"
	"github.com/knowhunger/ortoo/pkg/context"
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/pkg/types"
	"github.com/knowhunger/ortoo/server/constants"
	"reflect"
)

// ProcessPushPull processes a GRPC for Push-Pull
func (its *OrtooService) ProcessPushPull(goctx gocontext.Context, in *model.PushPullRequest) (*model.PushPullResponse, error) {
	ctx := context.New(goctx)
	collectionDoc, rpcErr := its.getCollectionDocWithRPCError(ctx, in.Header.GetCollection())
	if rpcErr != nil {
		return nil, rpcErr
	}

	ctx.SetNewLogger(context.SERVER, context.MakeTagInRPCProcess(constants.TagPushPull, collectionDoc.Num, in.Header.Cuid))
	ctx.L().Infof("RECV %v", in.ToString())

	clientDoc, err := its.mongo.GetClient(ctx, types.UIDtoString(in.Header.GetCuid()))
	if err != nil {
		return nil, errors.NewRPCError(err)
	}
	if clientDoc == nil {
		msg := fmt.Sprintf("client '%s'", in.Header.GetClient())
		return nil, errors.NewRPCError(errors.ServerNoResource.New(ctx.L(), msg))
	}
	if clientDoc.CollectionNum != collectionDoc.Num {
		msg := fmt.Sprintf("client '%s' accesses collection(%d)", clientDoc.GetClient(), collectionDoc.Num)
		return nil, errors.NewRPCError(errors.ServerNoPermission.New(ctx.L(), msg))
	}

	response := &model.PushPullResponse{
		Header: in.Header,
	}

	var chanList []<-chan *model.PushPullPack

	for _, ppp := range in.PushPullPacks {
		handler := newPushPullHandler(goctx, ppp, clientDoc, collectionDoc, its)
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
	return response, nil
}
