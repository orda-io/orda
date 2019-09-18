package mongodb

import (
	"context"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/server/mongodb/schema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

type baseCollection struct {
	name       string
	collection *mongo.Collection
}

func newCollection(collection *mongo.Collection) *baseCollection {
	return &baseCollection{
		collection: collection,
	}
}

//Create creates an empty collection by inserting a document and immediately deleting it.
func (c *baseCollection) create(ctx context.Context, docModel schema.MongoDBDoc) error {
	result, err := c.collection.InsertOne(ctx, bson.D{})
	if err != nil {
		return log.OrtooError(err, "fail to create collection:%s", c.name)
	}
	if _, err = c.collection.DeleteOne(ctx, filterByID(result.InsertedID)); err != nil {
		return log.OrtooError(err, "fail to delete inserted one")
	}
	log.Logger.Infof("Create collection:%s", c.collection.Name())
	if docModel != nil {
		if err := c.createIndex(ctx, docModel); err != nil {
			return log.OrtooError(err, "fail to create indexes")
		}

	}
	return nil
}

func (c *baseCollection) createIndex(ctx context.Context, docModel schema.MongoDBDoc) error {
	indexModel := docModel.GetIndexModel()
	if len(indexModel) > 0 {
		indexes, err := c.collection.Indexes().CreateMany(ctx, indexModel)
		if err != nil {
			return log.OrtooError(err, "fail to create indexes")
		}
		log.Logger.Infof("index is created: %v in %s", indexes, reflect.TypeOf(docModel))
	}
	return nil
}
