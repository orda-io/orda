package mongodb

import (
	"context"
	"errors"
	log "github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/server/mongodb/schema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type CollectionDatatypes struct {
	*baseCollection
}

func NewCollectionDatatypes(client *mongo.Client, datatype *mongo.Collection) *CollectionDatatypes {
	return &CollectionDatatypes{newCollection(client, datatype)}
}

func (c *CollectionDatatypes) GetDatatype(ctx context.Context, duid string) (*schema.DatatypeDoc, error) {
	result := c.collection.FindOne(ctx, filterByID(duid))
	if err := result.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, log.OrtooError(err)
	}
	var datatype schema.DatatypeDoc
	if err := result.Decode(&datatype); err != nil {
		return nil, log.OrtooError(err)
	}
	return &datatype, nil
}

func (c *CollectionDatatypes) GetDatatypeByKey(ctx context.Context, collectionNum uint32, key string) (*schema.DatatypeDoc, error) {
	result := c.collection.FindOne(ctx, bson.D{
		bson.E{Key: schema.DatatypeDocFields.CollectionNum, Value: collectionNum},
		bson.E{Key: schema.DatatypeDocFields.Key, Value: key},
	})
	if err := result.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, log.OrtooError(err)
	}
	var datatype schema.DatatypeDoc
	if err := result.Decode(&datatype); err != nil {
		return nil, log.OrtooError(err)
	}
	return &datatype, nil
}

func (c *CollectionDatatypes) UpdateDatatype(ctx context.Context, datatype *schema.DatatypeDoc) error {
	result, err := c.collection.UpdateOne(ctx, filterByID(datatype.DUID), datatype.ToUpdateBSON(), upsertOption)
	if err != nil {
		return log.OrtooError(err)
	}

	if result.ModifiedCount == 1 || result.UpsertedCount == 1 {
		return nil
	}
	return log.OrtooError(errors.New("fail to update datatype"))
}
