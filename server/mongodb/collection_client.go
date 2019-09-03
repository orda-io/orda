package mongodb

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
)

type CollectionClient struct {
	*Collection
}

func NewCollectionClient(ctx context.Context, collection *mongo.Collection) *CollectionClient {
	return &CollectionClient{NewCollection(ctx, collection)}
}
