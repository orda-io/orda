package mongodb

import (
	"context"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/server/mongodb/schema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

// MongoCollections is a bunch of collections used to provide
type MongoCollections struct {
	mongoClient *mongo.Client
	clients     *mongo.Collection
	counters    *mongo.Collection
	snapshots   *mongo.Collection
	datatypes   *mongo.Collection
	operations  *mongo.Collection
	collections *mongo.Collection
}

// Create creates an empty collection by inserting a document and immediately deleting it.
func (m *MongoCollections) create(ctx context.Context, collection *mongo.Collection, docModel schema.MongoDBDoc) error {
	result, err := collection.InsertOne(ctx, bson.D{})
	if err != nil {
		return log.OrtooErrorf(err, "fail to create collection:%s", collection.Name())
	}
	if _, err = collection.DeleteOne(ctx, schema.FilterByID(result.InsertedID)); err != nil {
		return log.OrtooErrorf(err, "fail to delete inserted one")
	}
	log.Logger.Infof("create collection:%s", collection.Name())
	if docModel != nil {
		if err := m.createIndex(ctx, collection, docModel); err != nil {
			return log.OrtooErrorf(err, "fail to create indexes")
		}

	}
	return nil
}

func (m *MongoCollections) createIndex(ctx context.Context, collection *mongo.Collection, docModel schema.MongoDBDoc) error {
	if docModel != nil {
		indexModel := docModel.GetIndexModel()
		if len(indexModel) > 0 {
			indexes, err := collection.Indexes().CreateMany(ctx, indexModel)
			if err != nil {
				return log.OrtooErrorf(err, "fail to create indexes")
			}
			log.Logger.Infof("index is created: %v in %s", indexes, reflect.TypeOf(docModel))
		}
	}
	return nil
}

func (m *MongoCollections) doTransaction(ctx context.Context, transactions func() error) error {
	session, err := m.mongoClient.StartSession()
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
