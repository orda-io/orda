package mongodb

import (
	"context"
	"github.com/knowhunger/ortoo/commons/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Collection struct {
	name       string
	collection *mongo.Collection
	ctx        context.Context
}

func NewCollection(ctx context.Context, collection *mongo.Collection) *Collection {
	return &Collection{
		ctx:        ctx,
		collection: collection,
	}
}

func (c *Collection) Create() error {
	result, err := c.collection.InsertOne(c.ctx, bson.D{})
	if err != nil {
		return log.OrtooError(err, "fail to create collection:%s", c.name)
	}
	log.Logger.Infof("%+v", result.InsertedID)
	result2, err := c.collection.DeleteOne(c.ctx, bson.D{bson.E{"_id", result.InsertedID}})
	log.Logger.Infof("%+v", result2)
	log.Logger.Infof("Create collection:%s", c.collection.Name())
	return nil
}
