package mongodb

import (
	"context"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/server/mongodb/schema"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

//CollectionCollections is a struct for Collections
type CollectionCollections struct {
	*baseCollection
}

//NewCollectionCollections creates a new CollectionCollections
func NewCollectionCollections(client *mongo.Client, collection *mongo.Collection) *CollectionCollections {
	return &CollectionCollections{
		newCollection(client, collection),
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

//InsertCollection inserts a collection document
func (c *CollectionCollections) InsertCollection(ctx context.Context, name string) (*schema.CollectionDoc, error) {

	session, err := c.client.StartSession()
	if err != nil {
		return nil, log.OrtooErrorf(err, "fail to start session")
	}

	if err := session.StartTransaction(); err != nil {
		return nil, log.OrtooErrorf(err, "fail to start transaction")
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		//if result, err = collection.UpdateOne(sc, bson.M{"_id": id}, update); err != nil {
		//	t.Fatal(err)
		//}
		//if result.MatchedCount != 1 || result.ModifiedCount != 1 {
		//	t.Fatal("replace failed, expected 1 but got", result.MatchedCount)
		//}
		collection := schema.CollectionDoc{
			Name:      name,
			CreatedAt: time.Now(),
		}
		c.collection.Aggregate(ctx, pip)

		if err = session.CommitTransaction(sc); err != nil {
			return log.OrtooErrorf(err, "fail to commit transaction")
		}
		return nil
	}); err != nil {
		return nil, log.OrtooErrorf(err, "fail to work with session")
	}
	session.EndSession(ctx)

	_, err := c.collection.InsertOne(ctx, collection)
	if err != nil {
		return nil, log.OrtooErrorf(err, "fail to insert collection")
	}
	return &collection, nil
}

//PurgeAllCollection ...
func (c *CollectionCollections) PurgeAllCollection(ctx context.Context, name string) {

}
