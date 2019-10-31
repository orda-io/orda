package mongodb

import (
	"context"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/server/mongodb/schema"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

// CollectionCollections is a struct for Collections
type CollectionCollections struct {
	*baseCollection
	counterCollection *CollectionCounters
}

// NewCollectionCollections creates a new CollectionCollections
func NewCollectionCollections(client *mongo.Client, counterCollection *CollectionCounters, collection *mongo.Collection) *CollectionCollections {
	return &CollectionCollections{
		newCollection(client, collection),
		counterCollection,
	}
}

// GetCollection gets a collectionDoc by the name
func (c *CollectionCollections) GetCollection(ctx context.Context, name string) (*schema.CollectionDoc, error) {
	sr := c.collection.FindOne(ctx, schema.FilterByID(name))
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

func (c *CollectionCollections) DeleteCollection(ctx context.Context, name string) error {
	result, err := c.collection.DeleteOne(ctx, schema.FilterByID(name))
	if err != nil {
		return log.OrtooError(err)
	}
	if result.DeletedCount == 1 {
		log.Logger.Infof("Collection is successfully removed:%s", name)
	}
	return nil
}

// InsertCollection inserts a collection document
func (c *CollectionCollections) InsertCollection(ctx context.Context, name string) (collection *schema.CollectionDoc, err error) {

	if err := c.doTransaction(ctx, func() error {
		num, err := c.counterCollection.GetNextCollectionNum(ctx)
		if err != nil {
			return log.OrtooErrorf(err, "fail to get next counter")
		}
		collection = &schema.CollectionDoc{
			Name:      name,
			Num:       num,
			CreatedAt: time.Now(),
		}
		_, err = c.collection.InsertOne(ctx, collection)
		if err != nil {
			return log.OrtooErrorf(err, "fail to insert collection")
		}
		log.Logger.Infof("Collection is inserted: %+v", collection)
		return nil
	}); err != nil {
		return nil, log.OrtooError(err)
	}
	return collection, nil

	//session, err := c.client.StartSession()
	//if err != nil {
	//	return nil, log.OrtooErrorf(err, "fail to start session")
	//}
	//
	//if err := session.StartTransaction(); err != nil {
	//	return nil, log.OrtooErrorf(err, "fail to start transaction")
	//}
	//
	//if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
	//	num, err := c.counterCollection.GetNextCollectionNum(ctx)
	//	if err != nil {
	//		return log.OrtooErrorf(err, "fail to get next counter")
	//	}
	//	collection = &schema.CollectionDoc{
	//		Name:      name,
	//		Num:       num,
	//		CreatedAt: time.Now(),
	//	}
	//	_, err = c.collection.InsertOne(ctx, collection)
	//	if err != nil {
	//		return log.OrtooErrorf(err, "fail to insert collection")
	//	}
	//	log.Logger.Infof("Collection is inserted: %+v", collection)
	//	if err = session.CommitTransaction(sc); err != nil {
	//		return log.OrtooErrorf(err, "fail to commit transaction")
	//	}
	//	return nil
	//}); err != nil {
	//	return nil, log.OrtooErrorf(err, "fail to work with session")
	//}
	//session.EndSession(ctx)
	//
	//return collection, nil
}

// PurgeAllCollection ...
func (c *CollectionCollections) PurgeAllCollection(ctx context.Context, name string) {

}
