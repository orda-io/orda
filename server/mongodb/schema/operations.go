package schema

import (
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"time"
)

// OperationDoc defines a document for operation, stored in MongoDB
type OperationDoc struct {
	ID            string    `bson:"_id"`
	DUID          string    `bson:"duid"`
	CollectionNum uint32    `bson:"colNum"`
	OpType        string    `bson:"type"`
	Sseq          uint64    `bson:"sseq"`
	Operation     []byte    `bson:"op"`
	CreatedAt     time.Time `bson:"createdAt"`
}

// OperationDocFields defines the fields of OperationDoc
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

// GetIndexModel returns the index models of the collection of OperationDoc
func (c *OperationDoc) GetIndexModel() []mongo.IndexModel {
	return []mongo.IndexModel{{
		Keys: bsonx.Doc{
			{OperationDocFields.DUID, bsonx.Int32(1)},
			{OperationDocFields.Sseq, bsonx.Int32(-1)},
		},
	}}
}
