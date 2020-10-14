package mongodb

import (
	"fmt"
	"github.com/knowhunger/ortoo/pkg/context"
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/server/mongodb/schema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

// GetLatestSnapshot gets the latest snapshot for the specified datatype.
func (its *MongoCollections) GetLatestSnapshot(
	ctx context.OrtooContext,
	collectionNum uint32,
	duid string,
) (*schema.SnapshotDoc, errors.OrtooError) {
	f := schema.GetFilter().
		AddFilterEQ(schema.SnapshotDocFields.CollectionNum, collectionNum).
		AddFilterEQ(schema.SnapshotDocFields.DUID, duid)
	opt := options.FindOne()
	opt.SetSort(bson.D{{schema.SnapshotDocFields.Sseq, 1}})
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
	ctx context.OrtooContext,
	collectionNum uint32,
	duid string,
	sseq uint64,
	meta []byte,
	snapshot string,
) errors.OrtooError {
	snap := schema.SnapshotDoc{
		ID:            fmt.Sprintf("%s:%d", duid, sseq),
		CollectionNum: collectionNum,
		DUID:          duid,
		Sseq:          sseq,
		Meta:          meta,
		Snapshot:      snapshot,
		CreatedAt:     time.Now(),
	}
	result, err := its.snapshots.InsertOne(ctx, snap.ToInsertBSON())
	if err != nil {
		return errors.ServerDBQuery.New(ctx.L(), err.Error())
	}
	if result.InsertedID == snap.ID {
		ctx.L().Infof("[MONGO] insert snapshot: %s", result.InsertedID)
	}
	return nil
}
