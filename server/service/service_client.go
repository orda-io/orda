package service

import (
	gocontext "context"
	"fmt"
	"github.com/orda-io/orda/client/pkg/context"
	"github.com/orda-io/orda/client/pkg/errors"
	"github.com/orda-io/orda/client/pkg/model"
	"github.com/orda-io/orda/server/admin"
	"time"

	"github.com/orda-io/orda/server/schema"

	"github.com/orda-io/orda/server/constants"
)

// ProcessClient processes ClientRequest and returns ClientResponse
func (its *OrdaService) ProcessClient(
	goCtx gocontext.Context,
	req *model.ClientMessage,
) (*model.ClientMessage, error) {
	ctx := context.NewOrdaContext(goCtx, constants.TagClient).
		UpdateCollectionTags(req.Collection, 0).
		UpdateClientTags(req.GetClientAlias(), req.GetCuid())
	if admin.IsAdminCUID(req.GetCuid()) {
		return nil, errors.NewRPCError(
			errors.ServerNoPermission.New(ctx.L(),
				fmt.Sprintf("not allowed CUID '%s'", req.GetCuid())))
	}

	collectionDoc, rpcErr := its.getCollectionDocWithRPCError(ctx, req.Collection)
	if rpcErr != nil {
		return nil, rpcErr
	}
	ctx.UpdateCollectionTags(collectionDoc.Name, collectionDoc.Num)

	clientDocFromReq := schema.ClientModelToBson(req.GetClient(), collectionDoc.Num)

	ctx.L().Infof("REQ[CLIE] %s %v %v", req.ToString(), len(req.Cuid), req.Cuid)

	clientDocFromDB, err := its.managers.Mongo.GetClient(ctx, clientDocFromReq.CUID)
	if err != nil {
		return nil, errors.NewRPCError(err)
	}
	if clientDocFromDB == nil {
		clientDocFromReq.CreatedAt = time.Now()
		ctx.L().Infof("create a new client:%+v", clientDocFromReq)
		if err := its.managers.Mongo.GetOrCreateRealCollection(ctx, req.Collection); err != nil {
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
	if err = its.managers.Mongo.UpdateClient(ctx, clientDocFromReq); err != nil {
		return nil, errors.NewRPCError(err)
	}
	ctx.L().Infof("RES[CLIE] %s", req.ToString())
	return req, nil
}
