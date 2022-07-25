package mongodb

import (
	"github.com/orda-io/orda/client/pkg/context"
	errors2 "github.com/orda-io/orda/client/pkg/errors"
	"github.com/orda-io/orda/server/schema"
	"reflect"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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
func (its *MongoCollections) createCollection(
	ctx context.OrdaContext,
	collection *mongo.Collection,
	docModel schema.MongoDBDoc,
) errors2.OrdaError {
	result, err := collection.InsertOne(ctx, bson.D{})
	if err != nil {
		return errors2.ServerDBQuery.New(ctx.L(), err.Error())
	}
	if _, err = collection.DeleteOne(ctx, schema.FilterByID(result.InsertedID)); err != nil {
		return errors2.ServerDBQuery.New(ctx.L(), err.Error())
	}
	ctx.L().Infof("create collection:%s", collection.Name())
	if docModel != nil {
		return its.createIndex(ctx, collection, docModel)
	}
	return nil
}

func (its *MongoCollections) createIndex(
	ctx context.OrdaContext,
	collection *mongo.Collection,
	docModel schema.MongoDBDoc,
) errors2.OrdaError {
	if docModel != nil {
		indexModel := docModel.GetIndexModel()
		if len(indexModel) > 0 {
			indexes, err := collection.Indexes().CreateMany(ctx, indexModel)
			if err != nil {
				return errors2.ServerDBQuery.New(ctx.L(), err.Error())
			}
			ctx.L().Infof("index is created: %v in %s", indexes, reflect.TypeOf(docModel))
		}
	}
	return nil
}

func (its *MongoCollections) doTransaction(
	ctx context.OrdaContext,
	transactions func() errors2.OrdaError,
) errors2.OrdaError {
	session, err := its.mongoClient.StartSession()
	if err != nil {
		return errors2.ServerDBQuery.New(ctx.L(), err.Error())
	}

	if err := session.StartTransaction(); err != nil {
		return errors2.ServerDBQuery.New(ctx.L(), err.Error())
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		if err := transactions(); err != nil {
			return err
		}
		if err := session.CommitTransaction(sc); err != nil {
			return errors2.ServerDBQuery.New(ctx.L(), err.Error())
		}
		return nil
	}); err != nil {
		return err.(errors2.OrdaError)
	}
	session.EndSession(ctx)
	return nil
}
