package mongodb

import (
	"context"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/server/mongodb/schema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type CollectionClient struct {
	*Collection
}

func NewCollectionClient(ctx context.Context, collection *mongo.Collection) *CollectionClient {
	return &CollectionClient{NewCollection(ctx, collection)}
}

func (c *CollectionClient) CreateClient(client *schema.ClientDoc) {
	client.CreatedAt = time.Now()
	client.UpdatedAt = time.Now()
	//b, err := bson.Marshal(client)
	//if err != nil {
	//	_ = log.OrtooError(err, "fail to marshal something")
	//}

	filter := bson.D{bson.E{schema.ID, client.Cuid}}
	var upsert = true
	//updateModel := mongo.NewUpdateOneModel().SetUpsert(true)
	u := bson.D{
		{"$set", bson.D{
			{"alias", client.Alias},
			{"collection", client.Collection},
			{"createdAt", client.CreatedAt},
		}},
		{"$currentDate", bson.D{
			{"updatedAt", true},
		}},
	}
	_, err := c.collection.UpdateOne(c.ctx, filter, u, &options.UpdateOptions{
		Upsert: &upsert,
	})

	if err != nil {
		_ = log.OrtooError(err, "fail to insert")
	}

}

func (c *CollectionClient) GetClient(cuid string) (*schema.ClientDoc, error) {
	filter := bson.D{bson.E{schema.ID, cuid}}
	sr := c.collection.FindOne(c.ctx, filter)
	if err := sr.Err(); err != nil {
		if sr.Err() == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, log.OrtooError(err, "fail to get client")
	}

	var client *schema.ClientDoc
	if err := sr.Decode(client); err != nil {
		return nil, log.OrtooError(err, "fail to decode clientDoc")
	}
	return client, nil
}
