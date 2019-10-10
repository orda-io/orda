package mongodb

import (
	"context"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/server/mongodb/schema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	idForCollection = "collectionID"
)

type CollectionCounters struct {
	*baseCollection
}

func NewCollectionCounters(client *mongo.Client, collection *mongo.Collection) *CollectionCounters {
	return &CollectionCounters{newCollection(client, collection)}
}

func (c *CollectionCounters) GetNextCollectionNum(ctx context.Context) (uint32, error) {
	opts := options.FindOneAndUpdate()
	opts.SetUpsert(true)
	var update = bson.M{
		"$inc": bson.M{"num": 1},
	}
	//_ = json.Unmarshal([]byte(`{ "$inc": {"num": 1}}`), &update)
	result := c.collection.FindOneAndUpdate(ctx, filterByID(idForCollection), update, opts)
	if err := result.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return 1, nil
		}
		return 0, log.OrtooErrorf(err, "fail to findAndUpdate")
	}
	var counterDoc = schema.CounterDoc{}
	if err := result.Decode(&counterDoc); err != nil {
		return 0, log.OrtooErrorf(err, "fail to decode counter")
	}
	return counterDoc.Num, nil
}
