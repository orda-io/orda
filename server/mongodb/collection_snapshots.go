package mongodb

import (
	"context"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/server/mongodb/schema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

// func (m *MongoCollections) InsertSnapshot(ctx context.Context, collectionNum uint32, duid stri)
