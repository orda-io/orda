package mongodb

import (
	"context"
	"errors"
	log "github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/server/mongodb/schema"
	"go.mongodb.org/mongo-driver/mongo"
)

type CollectionDatatypes struct {
	*baseCollection
}

func NewCollectionDatatypes(client *mongo.Client, datatypes *mongo.Collection) *CollectionDatatypes {
	return &CollectionDatatypes{newCollection(client, datatypes)}
}

func (c *CollectionDatatypes) GetDatatype(ctx context.Context, duid string) (*schema.DatatypeDoc, error) {
	f := schema.GetFilter().AddFilterEQ(schema.DatatypeDocFields.DUID, duid)
	result := c.collection.FindOne(ctx, f)
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
	f := schema.GetFilter().
		AddFilterEQ(schema.DatatypeDocFields.CollectionNum, collectionNum).
		AddFilterEQ(schema.DatatypeDocFields.Key, key)
	result := c.collection.FindOne(ctx, f)
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
	f := schema.GetFilter().AddFilterEQ(schema.DatatypeDocFields.DUID, datatype.DUID)
	result, err := c.collection.UpdateOne(ctx, f, datatype.ToUpdateBSON(), schema.UpsertOption)
	if err != nil {
		return log.OrtooError(err)
	}

	if result.ModifiedCount == 1 || result.UpsertedCount == 1 {
		return nil
	}
	return log.OrtooError(errors.New("fail to update datatype"))
}

func (c *CollectionDatatypes) PurgeDatatype(ctx context.Context, collNum uint32, duid string) error {

	return nil
}
