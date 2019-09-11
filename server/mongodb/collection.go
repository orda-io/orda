package mongodb

import (
	"context"
	"github.com/knowhunger/ortoo/commons/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Collection struct {
	name       string
	collection *mongo.Collection
}

func NewCollection(collection *mongo.Collection) *Collection {
	return &Collection{
		collection: collection,
	}
}

//Create creates an empty collection by inserting a document and immediately deleting it.
func (c *Collection) Create(ctx context.Context) error {
	result, err := c.collection.InsertOne(ctx, bson.D{})
	if err != nil {
		return log.OrtooError(err, "fail to create collection:%s", c.name)
	}
	if _, err = c.collection.DeleteOne(ctx, filterByID(result.InsertedID)); err != nil {
		return log.OrtooError(err, "fail to delete inserted one")
	}
	log.Logger.Infof("Create collection:%s", c.collection.Name())
	return nil
}
