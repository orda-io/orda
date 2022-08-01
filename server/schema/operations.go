package schema

import (
	"fmt"
	"github.com/orda-io/orda/client/pkg/model"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

// NewOperationDoc creates a new OperationDoc with the given parameters
func NewOperationDoc(op *model.Operation, duid string, sseq uint64, colNum uint32) *OperationDoc {

	return &OperationDoc{
		ID:            fmt.Sprintf("%s:%d", duid, sseq),
		DUID:          duid,
		CollectionNum: colNum,
		OpType:        op.OpType.String(),
		OpID: OpID{
			Era:     op.ID.Era,
			Lamport: op.ID.Lamport,
			CUID:    op.ID.CUID,
			Seq:     op.ID.Seq,
		},
		Sseq:      sseq,
		Body:      op.Body,
		CreatedAt: time.Now(),
	}

}

type OpID struct {
	Era     uint32 `bson:"era"`
	Lamport uint64 `bson:"lamport"`
	CUID    string `bson:"cuid"`
	Seq     uint64 `bson:"seq"`
}

// OperationDoc defines a document for operation, stored in MongoDB
type OperationDoc struct {
	ID            string    `bson:"_id"`
	DUID          string    `bson:"duid"`
	CollectionNum uint32    `bson:"colNum"`
	OpType        string    `bson:"type"`
	OpID          OpID      `bson:"id"`
	Sseq          uint64    `bson:"sseq"`
	Body          []byte    `bson:"body"`
	CreatedAt     time.Time `bson:"createdAt"`
}

// GetOperation returns a model.Operation by composing parameters of OperationDoc.
func (its *OperationDoc) GetOperation() *model.Operation {
	opID := &model.OperationID{
		Era:     its.OpID.Era,
		Lamport: its.OpID.Lamport,
		CUID:    its.OpID.CUID,
		Seq:     its.OpID.Seq,
	}
	return &model.Operation{
		ID:     opID,
		OpType: model.TypeOfOperation(model.TypeOfOperation_value[its.OpType]),
		Body:   its.Body,
	}
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
func (its *OperationDoc) GetIndexModel() []mongo.IndexModel {
	return []mongo.IndexModel{{
		Keys: bsonx.Doc{
			{OperationDocFields.DUID, bsonx.Int32(1)},
			{OperationDocFields.Sseq, bsonx.Int32(-1)},
		},
	}}
}
