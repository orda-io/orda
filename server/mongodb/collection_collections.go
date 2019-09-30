package mongodb

import (
	"context"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/server/mongodb/schema"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

//CollectionCollections is a struct for Collections
type CollectionCollections struct {
	*baseCollection
}

//NewCollectionCollections creates a new CollectionCollections
func NewCollectionCollections(collection *mongo.Collection) *CollectionCollections {
	return &CollectionCollections{
		newCollection(collection),
	}
}

//GetCollections gets a collectionDoc by the name
func (c *CollectionCollections) GetCollections(ctx context.Context, name string) (*schema.CollectionDoc, error) {
	sr := c.collection.FindOne(ctx, filterByID(name))
	if err := sr.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, log.OrtooErrorf(err, "fail to get collection")
	}
	var collection schema.CollectionDoc
	if err := sr.Decode(&collection); err != nil {
		return nil, log.OrtooErrorf(err, "fail to decode collectionDoc")
	}
	return &collection, nil
}

//InsertCollection inserts a collection document
func (c *CollectionCollections) InsertCollection(ctx context.Context, name string) (*schema.CollectionDoc, error) {
	collection := schema.CollectionDoc{
		Name:      name,
		CreatedAt: time.Now(),
	}
	_, err := c.collection.InsertOne(ctx, collection)
	if err != nil {
		return nil, log.OrtooErrorf(err, "fail to insert collection")
	}
	return &collection, nil
}

//PurgeAllCollection ...
func (c *CollectionCollections) PurgeAllCollection(ctx context.Context, name string) {

}
