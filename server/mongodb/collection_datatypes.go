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
	collectionOperations *CollectionOperations
}

func NewCollectionDatatypes(client *mongo.Client, datatypes *mongo.Collection, collectionOperations *CollectionOperations) *CollectionDatatypes {
	return &CollectionDatatypes{
		baseCollection:       newCollection(client, datatypes),
		collectionOperations: collectionOperations,
	}
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

func (c *CollectionDatatypes) PurgeDatatype(ctx context.Context, collectionNum uint32, key string) error {
	doc, err := c.GetDatatypeByKey(ctx, collectionNum, key)
	if err != nil {
		return log.OrtooError(err)
	}
	if doc == nil {
		log.Logger.Warnf("find no datatype to purge")
		return nil
	}
	if err := c.doTransaction(ctx, func() error {
		if err := c.collectionOperations.PurgeOperations(ctx, collectionNum, doc.DUID); err != nil {
			return log.OrtooError(err)
		}
		f := schema.GetFilter().AddFilterEQ(schema.DatatypeDocFields.DUID, doc.DUID)
		result, err := c.collection.DeleteOne(ctx, f)
		if err != nil {
			return log.OrtooError(err)
		}
		if result.DeletedCount == 1 {
			log.Logger.Infof("purged datatype `%s(%d)`", key, collectionNum)
			return nil
		}
		log.Logger.Warnf("deleted no datatypeDoc")
		return nil
	}); err != nil {
		return log.OrtooError(err)
	}

	return nil
}
