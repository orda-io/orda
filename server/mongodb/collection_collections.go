package mongodb

import (
	"context"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/server/mongodb/schema"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"time"
)

//CollectionCollections is a struct for Collections
type CollectionCollections struct {
	*baseCollection
	counterCollection *CollectionCounters
}

//NewCollectionCollections creates a new CollectionCollections
func NewCollectionCollections(client *mongo.Client, counterCollection *CollectionCounters, collection *mongo.Collection) *CollectionCollections {
	return &CollectionCollections{
		newCollection(client, collection),
		counterCollection,
	}
}

//GetCollection gets a collectionDoc by the name
func (c *CollectionCollections) GetCollection(ctx context.Context, name string) (*schema.CollectionDoc, error) {
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

func (c *CollectionCollections) DeleteCollection(ctx context.Context, name string) error {
	return nil
}

//InsertCollection inserts a collection document
func (c *CollectionCollections) InsertCollection(ctx context.Context, name string) (collection *schema.CollectionDoc, err error) {

	session, err := c.client.StartSession()
	if err != nil {
		return nil, log.OrtooErrorf(err, "fail to start session")
	}

	opt := &options.TransactionOptions{
		ReadConcern:    readconcern.Majority(),
		ReadPreference: nil,
		WriteConcern:   nil,
		MaxCommitTime:  nil,
	}
	if err := session.StartTransaction(opt); err != nil {
		return nil, log.OrtooErrorf(err, "fail to start transaction")
	}

	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		//var num uint32 = 0
		//filter := bson.D{}
		//opts := options.Find()
		//opts.SetSort(bson.D{{Key: "num", Value: -1}})
		//opts.SetLimit(1)
		//cur, err := c.collection.Find(ctx, filter, opts)
		//if err != nil {
		//	if err == mongo.ErrNoDocuments {
		//		num = 0
		//	}
		//	return log.OrtooErrorf(err, "fail to get")
		//} else {
		//	var prev schema.CollectionDoc
		//	for cur.Next(ctx) {
		//		if err := cur.Decode(&prev); err != nil {
		//			return log.OrtooErrorf(err, "fail to decode collection")
		//		}
		//		num = prev.Num
		//	}
		//}
		//defer cur.Close(ctx)
		num, err := c.counterCollection.NextCollectionNum(ctx)
		if err != nil {
			return log.OrtooErrorf(err, "fail to get next counter")
		}

		collection = &schema.CollectionDoc{
			Name:      name,
			Num:       num + 1,
			CreatedAt: time.Now(),
		}
		_, err = c.collection.InsertOne(ctx, collection)
		if err != nil {
			return log.OrtooErrorf(err, "fail to insert collection")
		}
		if err = session.CommitTransaction(sc); err != nil {
			return log.OrtooErrorf(err, "fail to commit transaction")
		}
		return nil
	}); err != nil {
		return nil, log.OrtooErrorf(err, "fail to work with session")
	}
	session.EndSession(ctx)
	log.Logger.Infof("Collection is inserted: %+v", collection)
	return collection, nil
}

//PurgeAllCollection ...
func (c *CollectionCollections) PurgeAllCollection(ctx context.Context, name string) {

}
