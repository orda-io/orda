package mongodb

import (
	"github.com/knowhunger/ortoo/pkg/context"
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/server/mongodb/schema"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

// GetCollection gets a collectionDoc with the specified name.
func (its *MongoCollections) GetCollection(ctx context.OrtooContext, name string) (*schema.CollectionDoc, errors.OrtooError) {
	sr := its.collections.FindOne(ctx, schema.FilterByID(name))
	if err := sr.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, errors.ServerDBQuery.New(ctx.L(), err.Error())
	}
	var collection schema.CollectionDoc
	if err := sr.Decode(&collection); err != nil {
		return nil, errors.ServerDBDecode.New(ctx.L(), err.Error())
	}
	return &collection, nil
}

// DeleteCollection deletes collections with the specified name.
func (its *MongoCollections) DeleteCollection(ctx context.OrtooContext, name string) errors.OrtooError {
	result, err := its.collections.DeleteOne(ctx, schema.FilterByID(name))
	if err != nil {
		return errors.ServerDBQuery.New(ctx.L(), err.Error())
	}
	if result.DeletedCount == 1 {
		ctx.L().Infof("Collection is successfully removed:%s", name)
	}
	return nil
}

// InsertCollection inserts a document for the specified collection.
func (its *MongoCollections) InsertCollection(
	ctx context.OrtooContext,
	name string,
) (collection *schema.CollectionDoc, err errors.OrtooError) {
	if err := its.doTransaction(ctx, func() errors.OrtooError {
		num, err := its.GetNextCollectionNum(ctx)
		if err != nil {
			return err
		}
		collection = &schema.CollectionDoc{
			Name:      name,
			Num:       num,
			CreatedAt: time.Now(),
		}
		_, err2 := its.collections.InsertOne(ctx, collection)
		if err2 != nil {
			return errors.ServerDBQuery.New(ctx.L(), err2.Error())
		}
		ctx.L().Infof("insert collection: %+v", collection)
		return nil
	}); err != nil {
		return nil, err
	}
	return collection, nil
}

// PurgeAllDocumentsOfCollection purges all data for the specified collection.
func (its *MongoCollections) PurgeAllDocumentsOfCollection(ctx context.OrtooContext, name string) errors.OrtooError {
	if err := its.doTransaction(ctx, func() errors.OrtooError {
		collectionDoc, err := its.GetCollection(ctx, name)
		if err != nil {
			return err
		}
		if collectionDoc == nil {
			return nil
		}
		ctx.L().Infof("purge collection '%s' (%d)", name, collectionDoc.Num)
		if err := its.purgeAllCollectionDatatypes(ctx, collectionDoc.Num); err != nil {
			return err
		}
		if err := its.purgeAllCollectionClients(ctx, collectionDoc.Num); err != nil {
			return err
		}
		filter := schema.GetFilter().AddFilterEQ(schema.CollectionDocFields.Name, name)

		result, err2 := its.collections.DeleteOne(ctx, filter)
		if err2 != nil {
			return errors.ServerDBQuery.New(ctx.L(), err2.Error())
		}
		if result.DeletedCount > 0 {
			ctx.L().Infof("delete collection `%s`", name)
			return nil
		}
		ctx.L().Warnf("delete no collection")
		return nil
	}); err != nil {
		return err
	}
	return nil
}
