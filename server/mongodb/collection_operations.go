package mongodb

import (
	"fmt"
	"github.com/orda-io/orda/client/pkg/errors"
	"github.com/orda-io/orda/client/pkg/iface"
	"github.com/orda-io/orda/client/pkg/model"
	"github.com/orda-io/orda/server/schema"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/orda-io/orda/server/constants"
)

// InsertOperations inserts operations into MongoDB
func (its *MongoCollections) InsertOperations(
	ctx iface.OrdaContext,
	operations []interface{},
) errors.OrdaError {
	if operations == nil {
		return nil
	}
	result, err := its.operations.InsertMany(ctx, operations)
	if err != nil {
		return errors.ServerDBQuery.New(ctx.L(), err.Error())
	}
	if len(result.InsertedIDs) != len(operations) {
		msg := fmt.Sprintf("the inserted operations (%d) are less than all the intended ones (%d)",
			len(result.InsertedIDs), len(operations))
		return errors.ServerDBQuery.New(ctx.L(), msg)
	}
	return nil
}

// DeleteOperation deletes operations for the specified sseq
func (its *MongoCollections) DeleteOperation(
	ctx iface.OrdaContext,
	duid string,
	sseq uint32,
) (int64, errors.OrdaError) {
	f := schema.GetFilter().
		AddFilterEQ(schema.OperationDocFields.DUID, duid).
		AddFilterEQ(schema.OperationDocFields.Sseq, sseq)
	result, err := its.operations.DeleteOne(ctx, f)
	if err != nil {
		return 0, errors.ServerDBQuery.New(ctx.L(), err.Error())
	}
	return result.DeletedCount, nil
}

// GetOperations gets operations of the specified range. For each operation, a given handler is called.
func (its *MongoCollections) GetOperations(
	ctx iface.OrdaContext,
	duid string,
	from, to uint64,
) (model.OpList, []uint64, errors.OrdaError) {
	f := schema.GetFilter().
		AddFilterEQ(schema.OperationDocFields.DUID, duid).
		AddFilterGTE(schema.OperationDocFields.Sseq, from)
	if to != constants.InfinitySseq {
		f.AddFilterLTE(schema.OperationDocFields.Sseq, to)
	}
	opt := options.Find()

	opt.SetSort(bson.D{{
		Key:   schema.OperationDocFields.Sseq,
		Value: 1,
	}})
	cursor, err := its.operations.Find(ctx, f, opt)
	if err != nil {
		return nil, nil, errors.ServerDBQuery.New(ctx.L(), err.Error())
	}
	var opList []*model.Operation
	var sseqList []uint64
	for cursor.Next(ctx) {
		var opDoc schema.OperationDoc
		if err := cursor.Decode(&opDoc); err != nil {
			return nil, nil, errors.ServerDBDecode.New(ctx.L(), err.Error())
		}
		opList = append(opList, opDoc.GetOperation())
		sseqList = append(sseqList, opDoc.Sseq)
	}
	return opList, sseqList, nil
}

// PurgeOperations purges operations for the specified datatype.
func (its *MongoCollections) PurgeOperations(ctx iface.OrdaContext, collectionNum int32, duid string) errors.OrdaError {
	f := schema.GetFilter().
		AddFilterEQ(schema.OperationDocFields.CollectionNum, collectionNum).
		AddFilterEQ(schema.OperationDocFields.DUID, duid)
	result, err := its.operations.DeleteMany(ctx, f)
	if err != nil {
		return errors.ServerDBQuery.New(ctx.L(), err.Error())
	}
	if result.DeletedCount > 0 {
		ctx.L().Infof("deleted %d operations of %s(%d)", result.DeletedCount, duid, collectionNum)
		return nil
	}
	ctx.L().Warnf("deleted no operations")
	return nil
}
