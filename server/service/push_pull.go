package service

import (
	gocontext "context"
	"fmt"
	"github.com/knowhunger/ortoo/pkg/context"
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/pkg/types"
	"reflect"
)

const tagPushPull = "REQ_PUSHPULL"

// ProcessPushPull processes a GRPC for Push-Pull
func (its *OrtooService) ProcessPushPull(goctx gocontext.Context, in *model.PushPullRequest) (*model.PushPullResponse, error) {

	ctx := context.NewWithTag(goctx, context.SERVER, tagPushPull, in.Header.GetClientSummary())
	ctx.L().Infof("receive %v", in.ToString())

	collectionDoc, rpcErr := its.getCollectionDocWithRPCError(ctx, in.Header.GetCollection())
	if rpcErr != nil {
		return nil, rpcErr
	}

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
		ID:     in.ID,
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
			// errors.PushPullMissingOps.New(ctx.L(), "fail to ") // TODO: return errors.
			// _ = log.OrtooErrorf(nil, "fail to run")
			continue
		} else {
			// ch := chanList[chosen]
			// log.Logger.Infof("%v %v", ch, msg)
			ppp := value.Interface().(*model.PushPullPack)
			response.PushPullPacks = append(response.PushPullPacks, ppp)
		}
	}
	return response, nil
}
