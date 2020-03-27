package schema

import (
	"fmt"
	"github.com/knowhunger/ortoo/ortoo/model"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"time"
)

// NewOperationDoc creates a new OperationDoc with the given parameters
func NewOperationDoc(op *model.Operation, duid string, sseq uint64, colNum uint32) *OperationDoc {

	return &OperationDoc{
		ID:            fmt.Sprintf("%s:%d", duid, sseq),
		DUID:          duid,
		CollectionNum: colNum,
		OpType:        op.OpType.String(),
		Era:           op.ID.Era,
		Lamport:       op.ID.Lamport,
		CUID:          op.ID.CUID,
		Seq:           op.ID.Seq,
		Sseq:          sseq,
		JSON:          string(op.Json),
		CreatedAt:     time.Now(),
	}

}

// OperationDoc defines a document for operation, stored in MongoDB
type OperationDoc struct {
	ID            string    `bson:"_id"`
	DUID          string    `bson:"duid"`
	CollectionNum uint32    `bson:"colNum"`
	OpType        string    `bson:"type"`
	Era           uint32    `bson:"era"`
	Lamport       uint64    `bson:"lamport"`
	CUID          []byte    `bson:"cuid"`
	Seq           uint64    `bson:"seq"`
	Sseq          uint64    `bson:"sseq"`
	JSON          string    `bson:"json"`
	CreatedAt     time.Time `bson:"createdAt"`
}

// GetOperation returns a model.Operation by composing parameters of OperationDoc.
func (its *OperationDoc) GetOperation() *model.Operation {
	opID := &model.OperationID{
		Era:     its.Era,
		Lamport: its.Lamport,
		CUID:    its.CUID,
		Seq:     its.Seq,
	}
	return &model.Operation{
		ID:     opID,
		OpType: model.TypeOfOperation(model.TypeOfOperation_value[its.OpType]),
		Json:   []byte(its.JSON),
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
