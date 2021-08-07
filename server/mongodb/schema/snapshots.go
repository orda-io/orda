package schema

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

// SnapshotDoc defines the document for snapshot, stored in MongoDB
type SnapshotDoc struct {
	ID            string    `bson:"_id"`
	CollectionNum uint32    `bson:"colNum"`
	DUID          string    `bson:"duid"`
	Sseq          uint64    `bson:"sseq"`
	Meta          string    `bson:"meta"`
	Snapshot      []byte    `bson:"snapshot"`
	CreatedAt     time.Time `bson:"createdAt"`
}

// SnapshotDocFields defines the fields of SnapshotDoc
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

// GetIndexModel returns the index models of the collection of SnapshotDoc
func (c *SnapshotDoc) GetIndexModel() []mongo.IndexModel {
	return []mongo.IndexModel{{
		Keys: bsonx.Doc{
			{SnapshotDocFields.CollectionNum, bsonx.Int32(1)},
			{SnapshotDocFields.DUID, bsonx.Int32(1)},
			{SnapshotDocFields.Sseq, bsonx.Int64(1)},
		},
	}}
}

// ToInsertBSON transforms SnapshotDoc to BSON type
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
