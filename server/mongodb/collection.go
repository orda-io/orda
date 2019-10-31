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
	client     *mongo.Client
	collection *mongo.Collection
}

func newCollection(client *mongo.Client, collection *mongo.Collection) *baseCollection {
	return &baseCollection{
		client:     client,
		collection: collection,
	}
}

//Create creates an empty collection by inserting a document and immediately deleting it.
func (c *baseCollection) create(ctx context.Context, docModel schema.MongoDBDoc) error {
	result, err := c.collection.InsertOne(ctx, bson.D{})
	if err != nil {
		return log.OrtooErrorf(err, "fail to create collection:%s", c.name)
	}
	if _, err = c.collection.DeleteOne(ctx, schema.FilterByID(result.InsertedID)); err != nil {
		return log.OrtooErrorf(err, "fail to delete inserted one")
	}
	log.Logger.Infof("Create collection:%s", c.collection.Name())
	if docModel != nil {
		if err := c.createIndex(ctx, docModel); err != nil {
			return log.OrtooErrorf(err, "fail to create indexes")
		}

	}
	return nil
}

func (c *baseCollection) createIndex(ctx context.Context, docModel schema.MongoDBDoc) error {
	indexModel := docModel.GetIndexModel()
	if len(indexModel) > 0 {
		indexes, err := c.collection.Indexes().CreateMany(ctx, indexModel)
		if err != nil {
			return log.OrtooErrorf(err, "fail to create indexes")
		}
		log.Logger.Infof("index is created: %v in %s", indexes, reflect.TypeOf(docModel))
	}
	return nil
}

func (c *baseCollection) doTransaction(ctx context.Context, transactions func() error) error {
	session, err := c.client.StartSession()
	if err != nil {
		return log.OrtooErrorf(err, "fail to start session")
	}

	if err := session.StartTransaction(); err != nil {
		return log.OrtooErrorf(err, "fail to start transaction")
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		if err := transactions(); err != nil {
			return log.OrtooError(err)
		}
		if err = session.CommitTransaction(sc); err != nil {
			return log.OrtooErrorf(err, "fail to commit transaction")
		}
		return nil
	}); err != nil {
		return log.OrtooErrorf(err, "fail to work with session")
	}
	session.EndSession(ctx)
	return nil
}
