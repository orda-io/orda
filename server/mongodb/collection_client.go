package mongodb

import "go.mongodb.org/mongo-driver/mongo"

type CollectionClient struct {
	*Collection
}

func NewCollectionClient(db *mongo.Database) *CollectionClient {
	return &CollectionClient{NewCollection(db)}
}
