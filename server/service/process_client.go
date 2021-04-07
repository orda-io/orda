package service

import (
	gocontext "context"
	"fmt"
	"github.com/knowhunger/ortoo/pkg/context"
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/server/constants"
	"github.com/knowhunger/ortoo/server/mongodb/schema"
	"time"
)

// ProcessClient processes ClientRequest and returns ClientResponse
func (its *OrtooService) ProcessClient(
	goCtx gocontext.Context,
	req *model.ClientMessage,
) (*model.ClientMessage, error) {
	ctx := context.New(goCtx)
	collectionDoc, rpcErr := its.getCollectionDocWithRPCError(ctx, req.GetClient().Collection)
	if rpcErr != nil {
		return nil, rpcErr
	}
	clientDocFromReq := schema.ClientModelToBson(req.GetClient(), collectionDoc.Num)

	ctx.SetNewLogger(context.SERVER, context.MakeTagInRPCProcess(constants.TagClient, collectionDoc.Num, req.GetClient().CUID))

	ctx.L().Infof("RECV %s %v %v", req.ToString(), len(req.Cuid), req.Cuid)

	clientDocFromDB, err := its.mongo.GetClient(ctx, clientDocFromReq.CUID)
	if err != nil {
		return nil, errors.NewRPCError(err)
	}
	if clientDocFromDB == nil {
		clientDocFromReq.CreatedAt = time.Now()
		ctx.L().Infof("create a new client:%+v", clientDocFromReq)
		if err := its.mongo.GetOrCreateRealCollection(ctx, req.GetClient().Collection); err != nil {
			return nil, errors.NewRPCError(err)
		}
	} else {
		if clientDocFromDB.CollectionNum != clientDocFromReq.CollectionNum {
			msg := fmt.Sprintf("client '%s' accesses collection(%d)",
				clientDocFromDB.GetClient(), clientDocFromReq.CollectionNum)
			return nil, errors.NewRPCError(errors.ServerNoPermission.New(ctx.L(), msg))
		}
		ctx.L().Infof("Client will be updated:%+v", clientDocFromReq)
	}
	clientDocFromReq.CreatedAt = time.Now()
	if err = its.mongo.UpdateClient(ctx, clientDocFromReq); err != nil {
		return nil, errors.NewRPCError(err)
	}
	ctx.L().Infof("SENDBACK %s", req.ToString())
	return req, nil
}
