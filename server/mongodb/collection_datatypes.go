package mongodb

import (
	"github.com/orda-io/orda/client/pkg/context"
	"github.com/orda-io/orda/client/pkg/errors"
	"github.com/orda-io/orda/server/schema"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetDatatype retrieves a datatypeDoc from MongoDB
func (its *MongoCollections) GetDatatype(
	ctx context.OrdaContext,
	duid string,
) (*schema.DatatypeDoc, errors.OrdaError) {
	f := schema.GetFilter().AddFilterEQ(schema.DatatypeDocFields.DUID, duid)
	result := its.datatypes.FindOne(ctx, f)
	if err := result.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, errors.ServerDBQuery.New(ctx.L(), err.Error())
	}
	var datatype schema.DatatypeDoc
	if err := result.Decode(&datatype); err != nil {
		return nil, errors.ServerDBDecode.New(ctx.L(), err.Error())
	}
	return &datatype, nil
}

// GetDatatypeByKey gets a datatype with the specified key.
func (its *MongoCollections) GetDatatypeByKey(
	ctx context.OrdaContext,
	collectionNum uint32,
	key string,
) (*schema.DatatypeDoc, errors.OrdaError) {
	f := schema.GetFilter().
		AddFilterEQ(schema.DatatypeDocFields.CollectionNum, collectionNum).
		AddFilterEQ(schema.DatatypeDocFields.Key, key)
	result := its.datatypes.FindOne(ctx, f)
	if err := result.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, errors.ServerDBQuery.New(ctx.L(), err.Error())
	}
	var datatype schema.DatatypeDoc
	if err := result.Decode(&datatype); err != nil {
		return nil, errors.ServerDBDecode.New(ctx.L(), err.Error())
	}
	return &datatype, nil
}

// UpdateDatatype updates the datatypeDoc.
func (its *MongoCollections) UpdateDatatype(
	ctx context.OrdaContext,
	datatype *schema.DatatypeDoc,
) errors.OrdaError {
	f := schema.GetFilter().AddFilterEQ(schema.DatatypeDocFields.DUID, datatype.DUID)
	result, err := its.datatypes.UpdateOne(ctx, f, datatype.ToUpdateBSON(), schema.UpsertOption)
	if err != nil {
		return errors.ServerDBQuery.New(ctx.L(), err.Error())
	}

	if result.ModifiedCount == 1 || result.UpsertedCount == 1 {
		return nil
	}
	return errors.ServerDBQuery.New(ctx.L(), "fail to update datatype")
}

func (its *MongoCollections) purgeAllCollectionDatatypes(
	ctx context.OrdaContext,
	collectionNum uint32,
) errors.OrdaError {
	opFilter := schema.GetFilter().AddFilterEQ(schema.OperationDocFields.CollectionNum, collectionNum)
	r1, err := its.operations.DeleteMany(ctx, opFilter)
	if err != nil {
		return errors.ServerDBQuery.New(ctx.L(), err.Error())
	}
	if r1.DeletedCount > 0 {
		ctx.L().Infof("delete %d operations in collection %d", r1.DeletedCount, collectionNum)
	}

	snapFilter := schema.GetFilter().AddFilterEQ(schema.SnapshotDocFields.CollectionNum, collectionNum)
	r2, err := its.snapshots.DeleteMany(ctx, snapFilter)
	if err != nil {
		return errors.ServerDBQuery.New(ctx.L(), err.Error())
	}
	if r2.DeletedCount > 0 {
		ctx.L().Infof("delete %d snapshots in collection %d", r2.DeletedCount, collectionNum)
	}

	datatypeFilter := schema.GetFilter().AddFilterEQ(schema.DatatypeDocFields.CollectionNum, collectionNum)
	r3, err := its.datatypes.DeleteMany(ctx, datatypeFilter)
	if err != nil {
		return errors.ServerDBQuery.New(ctx.L(), err.Error())
	}
	if r3.DeletedCount > 0 {
		ctx.L().Infof("delete %d datatypes in collection %d", r3.DeletedCount, collectionNum)
	}
	return nil
}

// PurgeDatatype purges a datatype from MongoDB.
func (its *MongoCollections) PurgeDatatype(
	ctx context.OrdaContext,
	collectionNum uint32,
	key string,
) errors.OrdaError {
	doc, err := its.GetDatatypeByKey(ctx, collectionNum, key)
	if err != nil {
		return err
	}
	if doc == nil {
		ctx.L().Warnf("find no datatype to purge")
		return nil
	}
	if err := its.doTransaction(ctx, func() errors.OrdaError {
		if err := its.PurgeOperations(ctx, collectionNum, doc.DUID); err != nil {
			return err
		}
		filter := schema.GetFilter().AddFilterEQ(schema.DatatypeDocFields.DUID, doc.DUID)
		result, err := its.datatypes.DeleteOne(ctx, filter)
		if err != nil {
			return errors.ServerDBQuery.New(ctx.L(), err.Error())
		}
		if result.DeletedCount == 1 {
			ctx.L().Infof("purged datatype `%s(%d)`", key, collectionNum)
			return nil
		}
		ctx.L().Warnf("deleted no datatypeDoc")
		return nil
	}); err != nil {
		return err
	}

	return nil
}
