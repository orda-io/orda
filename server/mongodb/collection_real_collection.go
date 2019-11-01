package mongodb

import (
	"go.mongodb.org/mongo-driver/mongo"
)

//CollectionSnapshots is used for manipulating snapshot of datatypes.
type CollectionSnapshots struct {
	*MongoCollections
	name string
}

func newCollectionSnapshot(client *mongo.Client, collection *mongo.Collection, name string) *CollectionSnapshots {
	return &CollectionSnapshots{
		MongoCollections: newCollection(client, collection),
		name:             name,
	}
}
