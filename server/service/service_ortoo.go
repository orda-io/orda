package service

import (
	"github.com/orda-io/orda/pkg/context"
	"github.com/orda-io/orda/pkg/errors"
	"github.com/orda-io/orda/server/mongodb"
	"github.com/orda-io/orda/server/notification"
	"github.com/orda-io/orda/server/schema"
)

// OrdaService is a rpc service of Orda
type OrdaService struct {
	mongo    *mongodb.RepositoryMongo
	notifier *notification.Notifier
}

// NewOrdaService creates a new OrdaService
func NewOrdaService(mongo *mongodb.RepositoryMongo, notifier *notification.Notifier) *OrdaService {
	return &OrdaService{
		mongo:    mongo,
		notifier: notifier,
	}
}

func (its *OrdaService) getCollectionDocWithRPCError(
	ctx context.OrdaContext,
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
