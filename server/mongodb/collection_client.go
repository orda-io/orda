package mongodb

import (
	"context"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/server/mongodb/schema"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type CollectionClient struct {
	*Collection
}

func NewCollectionClient(collection *mongo.Collection) *CollectionClient {
	return &CollectionClient{NewCollection(collection)}
}

func (c *CollectionClient) UpdateClient(ctx context.Context, client *schema.ClientDoc) {
	client.CreatedAt = time.Now()
	_, err := c.collection.UpdateOne(ctx, filterByID(client.Cuid), client.ToUpdateBson(), upsertOption)

	if err != nil {
		_ = log.OrtooError(err, "fail to insert")
	}

}

func (c *CollectionClient) GetClient(ctx context.Context, cuid string) (*schema.ClientDoc, error) {
	sr := c.collection.FindOne(ctx, filterByID(cuid))
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
