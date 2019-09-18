package mongodb

import (
	"go.mongodb.org/mongo-driver/mongo"
)

//CollectionSnapshots is used for manipulating snapshot of datatypes.
type CollectionSnapshots struct {
	*baseCollection
	name string
}

func newCollectionSnapshot(collection *mongo.Collection, name string) *CollectionSnapshots {

	return &CollectionSnapshots{
		baseCollection: newCollection(collection),
		name:           name,
	}
}
