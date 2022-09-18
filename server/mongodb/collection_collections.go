package mongodb

import (
	"github.com/orda-io/orda/client/pkg/errors"
	"github.com/orda-io/orda/client/pkg/iface"
	"github.com/orda-io/orda/server/schema"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

// GetCollection gets a collectionDoc with the specified name.
func (its *MongoCollections) GetCollection(ctx iface.OrdaContext, name string) (*schema.CollectionDoc, errors.OrdaError) {
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
func (its *MongoCollections) DeleteCollection(ctx iface.OrdaContext, name string) errors.OrdaError {
	result, err := its.collections.DeleteOne(ctx, schema.FilterByID(name))
	if err != nil {
		return errors.ServerDBQuery.New(ctx.L(), err.Error())
	}
	ctx.L().Infof("Collection '%s' is successfully removed:%d", name, result.DeletedCount)
	return nil
}

// InsertCollection inserts a document for the specified collection.
func (its *MongoCollections) InsertCollection(
	ctx iface.OrdaContext,
	name string,
) (collection *schema.CollectionDoc, err errors.OrdaError) {
	if err := its.doTransaction(ctx, func() errors.OrdaError {
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
func (its *MongoCollections) PurgeAllDocumentsOfCollection(ctx iface.OrdaContext, name string) errors.OrdaError {
	if err := its.doTransaction(ctx, func() errors.OrdaError {
		collectionDoc, err := its.GetCollection(ctx, name)
		if err != nil {
			return err
		}
		if collectionDoc == nil {
			return nil
		}
		ctx.L().Infof("purge collection#%d '%s'", collectionDoc.Num, name)
		return its.purgeAllDocumentsOfCollectionNum(ctx, collectionDoc.Num)
	}); err != nil {
		return err
	}
	return nil
}

// PurgeAllDocumentsOfCollection purges all data for the specified collection.
func (its *MongoCollections) purgeAllDocumentsOfCollectionNum(ctx iface.OrdaContext, collectionNum int32) errors.OrdaError {
	if err := its.purgeAllCollectionDatatypes(ctx, collectionNum); err != nil {
		return err
	}
	if err := its.purgeAllCollectionClients(ctx, collectionNum); err != nil {
		return err
	}
	filter := schema.GetFilter().AddFilterEQ(schema.CollectionDocFields.Name, collectionNum)

	result, err2 := its.collections.DeleteOne(ctx, filter)
	if err2 != nil {
		return errors.ServerDBQuery.New(ctx.L(), err2.Error())
	}
	ctx.L().Infof("delete %d collection#%d", result.DeletedCount, collectionNum)
	return nil

}
