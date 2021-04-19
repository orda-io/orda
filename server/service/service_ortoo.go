package service

import (
	"github.com/knowhunger/ortoo/pkg/context"
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/server/mongodb"
	"github.com/knowhunger/ortoo/server/mongodb/schema"
	"github.com/knowhunger/ortoo/server/notification"
)

// OrtooService is a rpc service of Ortoo
type OrtooService struct {
	mongo    *mongodb.RepositoryMongo
	notifier *notification.Notifier
}

// NewOrtooService creates a new OrtooService
func NewOrtooService(mongo *mongodb.RepositoryMongo, notifier *notification.Notifier) *OrtooService {
	return &OrtooService{
		mongo:    mongo,
		notifier: notifier,
	}
}

func (its *OrtooService) getCollectionDocWithRPCError(
	ctx context.OrtooContext,
	collection string,
) (*schema.CollectionDoc, error) {
	collectionDoc, err := its.mongo.GetCollection(ctx, collection)
	if err != nil {
		return nil, errors.NewRPCError(err)
	}
	if collectionDoc == nil {
		return nil, errors.NewRPCError(errors.ServerNoResource.New(ctx.L(), "collection "+collection))
	}
	return collectionDoc, nil
}
