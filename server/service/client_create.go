package service

import (
	"context"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
	"github.com/knowhunger/ortoo/server/mongodb/schema"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

//ProcessClient processes ClientRequest and returns ClientResponse
func (o *OrtooService) ProcessClient(ctx context.Context, in *model.ClientRequest) (*model.ClientResponse, error) {

	collectionDoc, err := o.mongo.GetCollection(ctx, in.Client.Collection)
	if err != nil {
		return nil, model.NewRPCError(model.RPCErrMongoDB)
	}
	if collectionDoc == nil {
		return nil, log.OrtooError(status.New(codes.InvalidArgument, "fail to find collection").Err())
	}

	transferredDoc := schema.ClientModelToBson(in.Client, collectionDoc.Num)

	storedDoc, err := o.mongo.GetClient(ctx, transferredDoc.CUID)
	if err != nil {
		return nil, log.OrtooErrorf(err, "fail to get client")
	}
	if storedDoc == nil {
		transferredDoc.CreatedAt = time.Now()
		log.Logger.Infof("A new client is created:%+v", transferredDoc)
		if err := o.mongo.GetOrCreateRealCollection(ctx, in.Client.Collection); err != nil {
			return nil, model.NewRPCError(model.RPCErrMongoDB)
		}
	} else {
		if storedDoc.CollectionNum != transferredDoc.CollectionNum {
			return nil, model.NewRPCError(model.RPCErrClientInconsistentCollection, storedDoc.CollectionNum, transferredDoc.CollectionNum)
		}
		log.Logger.Infof("Client will be updated:%+v", transferredDoc)
	}
	transferredDoc.CreatedAt = time.Now()
	if err = o.mongo.UpdateClient(ctx, transferredDoc); err != nil {
		return nil, model.NewRPCError(model.RPCErrMongoDB)
	}

	return model.NewClientResponse(in.Header, model.StateOfResponse_OK), nil
}
