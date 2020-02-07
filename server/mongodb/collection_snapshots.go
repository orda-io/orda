package mongodb

import (
	"context"
	"fmt"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/server/mongodb/schema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func (m *MongoCollections) GetLatestSnapshot(ctx context.Context, collectionNum uint32, duid string) (*schema.SnapshotDoc, error) {
	f := schema.GetFilter().
		AddFilterEQ(schema.SnapshotDocFields.CollectionNum, collectionNum).
		AddFilterEQ(schema.SnapshotDocFields.DUID, duid)
	opt := options.FindOne()
	opt.SetSort(bson.D{{schema.SnapshotDocFields.Sseq, 1}})
	result := m.snapshots.FindOne(ctx, f, opt)
	if err := result.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, log.OrtooError(err)
	}
	var snapshot schema.SnapshotDoc
	if err := result.Decode(&snapshot); err != nil {
		return nil, log.OrtooError(err)
	}
	return &snapshot, nil
}

func (m *MongoCollections) InsertSnapshot(ctx context.Context, collectionNum uint32, duid string, sseq uint64, meta []byte, snapshot string) error {
	snap := schema.SnapshotDoc{
		ID:            fmt.Sprintf("%s:%d", duid, sseq),
		CollectionNum: collectionNum,
		DUID:          duid,
		Sseq:          sseq,
		Meta:          meta,
		Snapshot:      snapshot,
		CreatedAt:     time.Now(),
	}
	result, err := m.snapshots.InsertOne(ctx, snap.ToInsertBSON())
	if err != nil {
		return log.OrtooError(err)
	}
	if result.InsertedID == snap.ID {
		log.Logger.Infof("Snapshot %s is inserted", result.InsertedID)
	}
	return nil
}
