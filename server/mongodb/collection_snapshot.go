package mongodb

import (
	"go.mongodb.org/mongo-driver/mongo"
)

type CollectionSnapshot struct {
	*Collection
	name string
}

func newCollectionSnapshot(collection *mongo.Collection, name string) *CollectionSnapshot {

	return &CollectionSnapshot{
		Collection: NewCollection(collection),
		name:       name,
	}
}
