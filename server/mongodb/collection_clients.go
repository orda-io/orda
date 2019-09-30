package mongodb

import (
	"context"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/server/mongodb/schema"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

//CollectionClients is a struct for Clients
type CollectionClients struct {
	*baseCollection
}

//NewCollectionClients creates a new CollectionClients
func NewCollectionClients(collection *mongo.Collection) *CollectionClients {
	return &CollectionClients{newCollection(collection)}
}

//UpdateClient updates a clientDoc; if not exists, a new clientDoc is inserted.
func (c *CollectionClients) UpdateClient(ctx context.Context, client *schema.ClientDoc) error {
	client.CreatedAt = time.Now()
	_, err := c.collection.UpdateOne(ctx, filterByID(client.Cuid), client.ToUpdateBson(), upsertOption)

	if err != nil {
		return log.OrtooErrorf(err, "fail to insert")
	}
	return nil
}

//GetClient gets a client with CUID
func (c *CollectionClients) GetClient(ctx context.Context, cuid string) (*schema.ClientDoc, error) {
	sr := c.collection.FindOne(ctx, filterByID(cuid))
	if err := sr.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, log.OrtooErrorf(err, "fail to get client")
	}

	var client *schema.ClientDoc
	if err := sr.Decode(client); err != nil {
		return nil, log.OrtooErrorf(err, "fail to decode clientDoc")
	}
	return client, nil
}
