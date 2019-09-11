package service

import (
	"context"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
	"github.com/knowhunger/ortoo/server/mongodb/schema"
	"time"
)

func (o *OrtooService) ClientCreate(ctx context.Context, in *model.ClientRequest) (*model.ClientReply, error) {

	transferredDoc := schema.ClientModelToBson(in.Client)
	storedDoc, err := o.mongo.GetClient(ctx, transferredDoc.Cuid)

	if err != nil {
		return nil, log.OrtooError(err, "fail to get client")
	}
	if storedDoc == nil {
		transferredDoc.CreatedAt = time.Now()
		log.Logger.Infof("A new client is created:%+v", transferredDoc)
		if _, err := o.mongo.GetOrCreateCollectionSnapshot(ctx, transferredDoc.Collection); err != nil {
			return nil, model.NewRPCError(model.RPCErrMongoDB)
		}
	} else {
		if storedDoc.Collection != transferredDoc.Collection {
			return nil, model.NewRPCError(model.RPCErrMongoDB, storedDoc.Collection, transferredDoc.Collection)
		}
	}
	o.mongo.UpdateClient(ctx, transferredDoc)
	return model.NewClientReply(in.Header.Seq, model.TypeReplyStates_OK), nil
}
