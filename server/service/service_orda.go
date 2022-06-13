package service

import (
	"github.com/orda-io/orda/pkg/context"
	"github.com/orda-io/orda/pkg/errors"
	"github.com/orda-io/orda/server/managers"
	"github.com/orda-io/orda/server/schema"
)

// OrdaService is a rpc service of Orda
type OrdaService struct {
	managers *managers.Managers
}

// NewOrdaService creates a new OrdaService
func NewOrdaService(managers *managers.Managers) *OrdaService {
	return &OrdaService{
		managers: managers,
	}
}

func (its *OrdaService) getCollectionDocWithRPCError(
	ctx context.OrdaContext,
	collection string,
) (*schema.CollectionDoc, error) {
	collectionDoc, err := its.managers.Mongo.GetCollection(ctx, collection)
	if err != nil {
		return nil, errors.NewRPCError(err)
	}
	if collectionDoc == nil {
		return nil, errors.NewRPCError(errors.ServerNoResource.New(ctx.L(), "collection "+collection))
	}
	return collectionDoc, nil
}
