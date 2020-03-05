package mongodb

import (
	"context"
	"errors"
	log "github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/server/mongodb/schema"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetDatatype retrieves a datatypeDoc from MongoDB
func (m *MongoCollections) GetDatatype(ctx context.Context, duid string) (*schema.DatatypeDoc, error) {
	f := schema.GetFilter().AddFilterEQ(schema.DatatypeDocFields.DUID, duid)
	result := m.datatypes.FindOne(ctx, f)
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

// GetDatatypeByKey gets a datatype with the specified key.
func (m *MongoCollections) GetDatatypeByKey(ctx context.Context, collectionNum uint32, key string) (*schema.DatatypeDoc, error) {
	f := schema.GetFilter().
		AddFilterEQ(schema.DatatypeDocFields.CollectionNum, collectionNum).
		AddFilterEQ(schema.DatatypeDocFields.Key, key)
	result := m.datatypes.FindOne(ctx, f)
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

// UpdateDatatype updates the datatypeDoc.
func (m *MongoCollections) UpdateDatatype(ctx context.Context, datatype *schema.DatatypeDoc) error {
	f := schema.GetFilter().AddFilterEQ(schema.DatatypeDocFields.DUID, datatype.DUID)
	result, err := m.datatypes.UpdateOne(ctx, f, datatype.ToUpdateBSON(), schema.UpsertOption)
	if err != nil {
		return log.OrtooError(err)
	}

	if result.ModifiedCount == 1 || result.UpsertedCount == 1 {
		return nil
	}
	return log.OrtooError(errors.New("fail to update datatype"))
}

func (m *MongoCollections) purgeAllCollectionDatatypes(ctx context.Context, collectionNum uint32) error {
	opFilter := schema.GetFilter().AddFilterEQ(schema.OperationDocFields.CollectionNum, collectionNum)
	r1, err := m.operations.DeleteMany(ctx, opFilter)
	if err != nil {
		return log.OrtooError(err)
	}
	if r1.DeletedCount > 0 {
		log.Logger.Infof("delete %d operations in collection %d", r1.DeletedCount, collectionNum)
	}

	snapFilter := schema.GetFilter().AddFilterEQ(schema.SnapshotDocFields.CollectionNum, collectionNum)
	r2, err := m.snapshots.DeleteMany(ctx, snapFilter)
	if err != nil {
		return log.OrtooError(err)
	}
	if r2.DeletedCount > 0 {
		log.Logger.Infof("delete %d snapshots in collection %d", r2.DeletedCount, collectionNum)
	}

	datatypeFilter := schema.GetFilter().AddFilterEQ(schema.DatatypeDocFields.CollectionNum, collectionNum)
	r3, err := m.datatypes.DeleteMany(ctx, datatypeFilter)
	if err != nil {
		return log.OrtooError(err)
	}
	if r3.DeletedCount > 0 {
		log.Logger.Infof("delete %d datatypes in collection %d", r3.DeletedCount, collectionNum)
	}
	return nil
}

// PurgeDatatype purges a datatype from MongoDB.
func (m *MongoCollections) PurgeDatatype(ctx context.Context, collectionNum uint32, key string) error {
	doc, err := m.GetDatatypeByKey(ctx, collectionNum, key)
	if err != nil {
		return log.OrtooError(err)
	}
	if doc == nil {
		log.Logger.Warnf("find no datatype to purge")
		return nil
	}
	if err := m.doTransaction(ctx, func() error {

		if err := m.PurgeOperations(ctx, collectionNum, doc.DUID); err != nil {
			return log.OrtooError(err)
		}
		filter := schema.GetFilter().AddFilterEQ(schema.DatatypeDocFields.DUID, doc.DUID)
		result, err := m.datatypes.DeleteOne(ctx, filter)
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
