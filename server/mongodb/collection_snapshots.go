package mongodb

import (
	"fmt"
	"github.com/orda-io/orda/server/schema"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/orda-io/orda/pkg/context"
	"github.com/orda-io/orda/pkg/errors"
)

// GetLatestSnapshot gets the latest snapshot for the specified datatype.
func (its *MongoCollections) GetLatestSnapshot(
	ctx context.OrdaContext,
	collectionNum uint32,
	duid string,
) (*schema.SnapshotDoc, errors.OrdaError) {
	f := schema.GetFilter().
		AddFilterEQ(schema.SnapshotDocFields.CollectionNum, collectionNum).
		AddFilterEQ(schema.SnapshotDocFields.DUID, duid)
	opt := options.FindOne()
	opt.SetSort(bson.D{{schema.SnapshotDocFields.Sseq, -1}})
	result := its.snapshots.FindOne(ctx, f, opt)
	if err := result.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, errors.ServerDBQuery.New(ctx.L(), err.Error())
	}
	var snapshot schema.SnapshotDoc
	if err := result.Decode(&snapshot); err != nil {
		return nil, errors.ServerDBDecode.New(ctx.L(), err.Error())
	}
	return &snapshot, nil
}

// InsertSnapshot inserts a snapshot for the specified datatype.
func (its *MongoCollections) InsertSnapshot(
	ctx context.OrdaContext,
	collectionNum uint32,
	duid string,
	sseq uint64,
	meta []byte,
	snapshot []byte,
) errors.OrdaError {
	snap := schema.SnapshotDoc{
		ID:            fmt.Sprintf("%s:%d", duid, sseq),
		CollectionNum: collectionNum,
		DUID:          duid,
		Sseq:          sseq,
		Meta:          string(meta),
		Snapshot:      snapshot,
		CreatedAt:     time.Now(),
	}
	result, err := its.snapshots.InsertOne(ctx, snap.ToInsertBSON())
	if err != nil {
		return errors.ServerDBQuery.New(ctx.L(), err.Error())
	}
	if result.InsertedID == snap.ID {
		ctx.L().Infof("insert snapshot: %s", result.InsertedID)
	}
	return nil
}
