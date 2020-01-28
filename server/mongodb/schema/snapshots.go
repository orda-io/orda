package schema

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"time"
)

type SnapshotDoc struct {
	ID            string    `bson:"_id"`
	CollectionNum uint32    `bson:"colNum"`
	DUID          string    `bson:"duid"`
	Sseq          uint64    `bson:"sseq"`
	Meta          []byte    `bson:"meta"`
	Snapshot      string    `bson:"snapshot"`
	CreatedAt     time.Time `bson:"createdAt"`
}

var SnapshotDocFields = struct {
	ID            string
	CollectionNum string
	DUID          string
	Sseq          string
	Meta          string
	Snapshot      string
	CreatedAt     string
}{
	ID:            "_id",
	CollectionNum: "colNum",
	DUID:          "duid",
	Sseq:          "sseq",
	Meta:          "meta",
	Snapshot:      "snapshot",
	CreatedAt:     "createdAt",
}

func (c *SnapshotDoc) GetIndexModel() []mongo.IndexModel {
	return []mongo.IndexModel{{
		Keys: bsonx.Doc{
			{SnapshotDocFields.CollectionNum, bsonx.Int32(1)},
			{SnapshotDocFields.DUID, bsonx.Int32(1)},
			{SnapshotDocFields.Sseq, bsonx.Int64(1)},
		},
	}}
}

func (c *SnapshotDoc) ToInsertBSON() bson.M {
	return bson.M{
		SnapshotDocFields.ID:            c.ID,
		SnapshotDocFields.CollectionNum: c.CollectionNum,
		SnapshotDocFields.DUID:          c.DUID,
		SnapshotDocFields.Sseq:          c.Sseq,
		SnapshotDocFields.Meta:          c.Meta,
		SnapshotDocFields.Snapshot:      c.Snapshot,
		SnapshotDocFields.CreatedAt:     c.CreatedAt,
	}
}
