package service

import (
	gocontext "context"
	"fmt"
	"github.com/knowhunger/ortoo/pkg/context"
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/server/mongodb/schema"
	"time"
)

const tagClient = "REQ_CLIENT"

// ProcessClient processes ClientRequest and returns ClientResponse
func (its *OrtooService) ProcessClient(goCtx gocontext.Context, in *model.ClientRequest) (*model.ClientResponse, error) {
	ctx := context.NewWithTag(goCtx, context.SERVER, tagClient, in.Client.GetSummary())
	ctx.L().Infof("receive %s", in.ToString())

	collectionDoc, rpcErr := its.getCollectionDocWithRPCError(ctx, in.Client.Collection)
	if rpcErr != nil {
		return nil, rpcErr
	}
	clientDocFromReq := schema.ClientModelToBson(in.Client, collectionDoc.Num)

	clientDocFromDB, err := its.mongo.GetClient(ctx, clientDocFromReq.CUID)
	if err != nil {
		return nil, errors.NewRPCError(err)
	}
	if clientDocFromDB == nil {
		clientDocFromReq.CreatedAt = time.Now()
		ctx.L().Infof("create a new client:%+v", clientDocFromReq)
		if err := its.mongo.GetOrCreateRealCollection(ctx, in.Client.Collection); err != nil {
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

	return model.NewClientResponse(in.Header, model.StateOfResponse_OK), nil
}
