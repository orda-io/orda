package service

import (
	gocontext "context"
	"fmt"
	"time"

	"github.com/orda-io/orda/pkg/errors"
	"github.com/orda-io/orda/pkg/model"
	"github.com/orda-io/orda/server/constants"
	"github.com/orda-io/orda/server/mongodb/schema"
	"github.com/orda-io/orda/server/svrcontext"
)

// ProcessClient processes ClientRequest and returns ClientResponse
func (its *OrdaService) ProcessClient(
	goCtx gocontext.Context,
	req *model.ClientMessage,
) (*model.ClientMessage, error) {
	ctx := svrcontext.NewServerContext(goCtx, constants.TagClient).
		UpdateCollection(req.Collection).
		UpdateClient(req.GetClientSummary())
	collectionDoc, rpcErr := its.getCollectionDocWithRPCError(ctx, req.Collection)
	if rpcErr != nil {
		return nil, rpcErr
	}
	ctx.UpdateCollection(collectionDoc.GetSummary())
	clientDocFromReq := schema.ClientModelToBson(req.GetClient(), collectionDoc.Num)

	ctx.L().Infof("REQ[CLIE] %s %v %v", req.ToString(), len(req.Cuid), req.Cuid)

	clientDocFromDB, err := its.mongo.GetClient(ctx, clientDocFromReq.CUID)
	if err != nil {
		return nil, errors.NewRPCError(err)
	}
	if clientDocFromDB == nil {
		clientDocFromReq.CreatedAt = time.Now()
		ctx.L().Infof("create a new client:%+v", clientDocFromReq)
		if err := its.mongo.GetOrCreateRealCollection(ctx, req.Collection); err != nil {
			return nil, errors.NewRPCError(err)
		}
	} else {

		if clientDocFromDB.CollectionNum != clientDocFromReq.CollectionNum {
			msg := fmt.Sprintf("client '%s' accesses collection(%d)",
				clientDocFromDB.ToString(), clientDocFromReq.CollectionNum)
			return nil, errors.NewRPCError(errors.ServerNoPermission.New(ctx.L(), msg))
		}
		ctx.L().Infof("Client will be updated:%+v", clientDocFromReq)
	}
	clientDocFromReq.CreatedAt = time.Now()
	if err = its.mongo.UpdateClient(ctx, clientDocFromReq); err != nil {
		return nil, errors.NewRPCError(err)
	}
	ctx.L().Infof("RES[CLIE] %s", req.ToString())
	return req, nil
}
