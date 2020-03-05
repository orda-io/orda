package mongodb

import (
	"context"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/server/constants"
	"github.com/knowhunger/ortoo/server/mongodb/schema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// InsertOperations inserts operations into MongoDB
func (m *MongoCollections) InsertOperations(ctx context.Context, operations []interface{}) error {
	if operations == nil {
		return nil
	}
	result, err := m.operations.InsertMany(ctx, operations)
	if err != nil {
		return log.OrtooError(err)
	}
	if len(result.InsertedIDs) != len(operations) {
		return log.OrtooErrorf(err, "fail to insert operation all")
	}
	return nil
}

// DeleteOperation deletes operations for the specified sseq
func (m *MongoCollections) DeleteOperation(ctx context.Context, duid string, sseq uint32) (int64, error) {
	f := schema.GetFilter().
		AddFilterEQ(schema.OperationDocFields.DUID, duid).
		AddFilterEQ(schema.OperationDocFields.Sseq, sseq)
	result, err := m.operations.DeleteOne(ctx, f)
	if err != nil {
		return 0, log.OrtooError(err)
	}
	return result.DeletedCount, nil
}

// GetOperations gets operations of the specified range. For each operation, a given handler is called.
func (m *MongoCollections) GetOperations(
	ctx context.Context,
	duid string,
	from, to uint64,
	operationDocHandler func(opDoc *schema.OperationDoc) error) error {
	f := schema.GetFilter().
		AddFilterEQ(schema.OperationDocFields.DUID, duid).
		AddFilterGTE(schema.OperationDocFields.Sseq, from)
	if to != constants.InfinitySseq {
		f.AddFilterLTE(schema.OperationDocFields.Sseq, to)
	}
	opt := options.Find()
	opt.SetSort(bson.D{{schema.OperationDocFields.Sseq, 1}})
	cursor, err := m.operations.Find(ctx, f, opt)
	if err != nil {
		return log.OrtooError(err)
	}

	for cursor.Next(ctx) {
		var operationDoc schema.OperationDoc
		if err := cursor.Decode(&operationDoc); err != nil {
			return log.OrtooError(err)
		}
		if operationDocHandler != nil {
			if err := operationDocHandler(&operationDoc); err != nil {
				return log.OrtooError(err)
			}
		}
	}
	return nil
}

// PurgeOperations purges operations for the specified datatype.
func (m *MongoCollections) PurgeOperations(ctx context.Context, collectionNum uint32, duid string) error {
	f := schema.GetFilter().
		AddFilterEQ(schema.OperationDocFields.CollectionNum, collectionNum).
		AddFilterEQ(schema.OperationDocFields.DUID, duid)
	result, err := m.operations.DeleteMany(ctx, f)
	if err != nil {
		return log.OrtooError(err)
	}
	if result.DeletedCount > 0 {
		log.Logger.Infof("deleted %d operations of %s(%d)", result.DeletedCount, duid, collectionNum)
		return nil
	}
	log.Logger.Warnf("deleted no operations")
	return nil
}
