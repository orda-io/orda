package mongodb

import (
	"context"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/server/mongodb/schema"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

// GetCollection gets a collectionDoc by the name
func (m *MongoCollections) GetCollection(ctx context.Context, name string) (*schema.CollectionDoc, error) {
	sr := m.collections.FindOne(ctx, schema.FilterByID(name))
	if err := sr.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, log.OrtooError(err)
	}
	var collection schema.CollectionDoc
	if err := sr.Decode(&collection); err != nil {
		return nil, log.OrtooError(err)
	}
	return &collection, nil
}

func (m *MongoCollections) DeleteCollection(ctx context.Context, name string) error {
	result, err := m.collections.DeleteOne(ctx, schema.FilterByID(name))
	if err != nil {
		return log.OrtooError(err)
	}
	if result.DeletedCount == 1 {
		log.Logger.Infof("Collection is successfully removed:%s", name)
	}
	return nil
}

// InsertCollection inserts a collection document
func (m *MongoCollections) InsertCollection(ctx context.Context, name string) (collection *schema.CollectionDoc, err error) {

	if err := m.doTransaction(ctx, func() error {
		num, err := m.GetNextCollectionNum(ctx)
		if err != nil {
			return log.OrtooErrorf(err, "fail to get next counter")
		}
		collection = &schema.CollectionDoc{
			Name:      name,
			Num:       num,
			CreatedAt: time.Now(),
		}
		_, err = m.collections.InsertOne(ctx, collection)
		if err != nil {
			return log.OrtooErrorf(err, "fail to insert collection")
		}
		log.Logger.Infof("insert collection: %+v", collection)
		return nil
	}); err != nil {
		return nil, log.OrtooError(err)
	}
	return collection, nil
}

// PurgeAllCollection ...
func (m *MongoCollections) PurgeAllCollection(ctx context.Context, name string) error {
	if err := m.doTransaction(ctx, func() error {
		collectionDoc, err := m.GetCollection(ctx, name)
		if err != nil {
			return log.OrtooError(err)
		}
		if collectionDoc == nil {
			return nil
		}
		log.Logger.Infof("purge collection '%s' (%d)", name, collectionDoc.Num)
		if err := m.PurgeAllCollectionDatatypes(ctx, collectionDoc.Num); err != nil {
			return log.OrtooError(err)
		}
		if err := m.PurgeAllCollectionClients(ctx, collectionDoc.Num); err != nil {
			return log.OrtooError(err)
		}
		filter := schema.GetFilter().AddFilterEQ(schema.CollectionDocFields.Name, name)
		result, err := m.collections.DeleteOne(ctx, filter)
		if err != nil {
			return log.OrtooError(err)
		}
		if result.DeletedCount > 0 {
			log.Logger.Infof("delete collection `%s`", name)
			return nil
		}
		log.Logger.Warnf("delete no collection")
		return nil
	}); err != nil {
		return log.OrtooError(err)
	}
	return nil
}
