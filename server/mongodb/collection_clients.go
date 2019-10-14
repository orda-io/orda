package mongodb

import (
	"context"
	"errors"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/server/mongodb/schema"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

//CollectionClients is a struct for Clients
type CollectionClients struct {
	*baseCollection
}

//NewCollectionClients creates a new CollectionClients
func NewCollectionClients(client *mongo.Client, collection *mongo.Collection) *CollectionClients {
	return &CollectionClients{newCollection(client, collection)}
}

//UpdateClient updates a clientDoc; if not exists, a new clientDoc is inserted.
func (c *CollectionClients) UpdateClient(ctx context.Context, client *schema.ClientDoc) error {
	client.CreatedAt = time.Now()
	res, err := c.collection.UpdateOne(ctx, filterByID(client.CUID), client.ToUpdateBson(), upsertOption)

	if err != nil {
		return log.OrtooErrorf(err, "fail to insert")
	}
	if res.ModifiedCount == 1 || res.UpsertedCount == 1 {
		return nil
	}
	return errors.New("fail to modify")
}

func (c *CollectionClients) DeleteClient(ctx context.Context, cuid string) error {
	result, err := c.collection.DeleteOne(ctx, filterByID(cuid))
	if err != nil {
		return log.OrtooError(err)
	}
	if result.DeletedCount == 1 {
		return nil
	}
	return log.OrtooError(errors.New("fail to find something to delete"))
}

//GetClient gets a client with CUID
func (c *CollectionClients) GetClient(ctx context.Context, cuid string) (*schema.ClientDoc, error) {
	opts := options.FindOne()
	//opts.SetProjection(bson.M{"checkPoint":0})
	sr := c.collection.FindOne(ctx, filterByID(cuid), opts)
	if err := sr.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, log.OrtooErrorf(err, "fail to get client")
	}

	var client schema.ClientDoc
	if err := sr.Decode(&client); err != nil {
		return nil, log.OrtooErrorf(err, "fail to decode clientDoc")
	}
	return &client, nil
}

func (c *CollectionClients) GetClientWithCheckPoint(ctx context.Context, cuid string, duid string) (*schema.ClientDoc, error) {
	//filter := bson.D{bson.E{Key: ID, Value: cuid}, bson.E{}}
	//
	//c.collection.FindOne(ctx, )
	return nil, nil
}
