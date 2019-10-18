package schema

import (
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"time"
)

type OperationDoc struct {
	ID            string      `bson:"_id"`
	DUID          string      `bson:"duid"`
	CollectionNum uint32      `bson:"colNum"`
	OpType        string      `bson:"type"`
	Sseq          uint64      `bson:"sseq"`
	Operation     interface{} `bson:"op"`
	CreatedAt     time.Time   `bson:"createdAt"`
}

var OperationDocFields = struct {
	ID            string
	DUID          string
	CollectionNum string
	OpType        string
	Sseq          string
	Operation     string
	CreatedAt     string
}{
	ID:            "_id",
	DUID:          "duid",
	CollectionNum: "colNum",
	OpType:        "type",
	Sseq:          "sseq",
	Operation:     "op",
	CreatedAt:     "createdAt",
}

func (c *OperationDoc) GetIndexModel() []mongo.IndexModel {
	return []mongo.IndexModel{{
		Keys: bsonx.Doc{
			{OperationDocFields.DUID, bsonx.Int32(1)},
			{OperationDocFields.Sseq, bsonx.Int32(-1)},
		},
	}}
}
