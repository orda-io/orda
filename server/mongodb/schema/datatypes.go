package schema

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"time"
)

type DatatypeDocState string

type DatatypeDoc struct {
	DUID          string    `bson:"_id"`
	Key           string    `bson:"key"`
	CollectionNum uint32    `bson:"colNum"`
	Type          string    `bson:"type"`
	Sseq          uint64    `bson:"sseq"`
	Visible       bool      `bson:"visible"`
	CreatedAt     time.Time `bson:"createdAt"`
	UpdatedAt     time.Time `bson:"updatedAt"`
}

var DatatypeDocFields = struct {
	DUID          string
	Key           string
	CollectionNum string
	Type          string
	Sseq          string
	CreatedAt     string
	UpdatedAt     string
}{
	DUID:          "_id",
	Key:           "key",
	CollectionNum: "colNum",
	Type:          "type",
	Sseq:          "sseq",
	CreatedAt:     "createdAt",
	UpdatedAt:     "updatedAt",
}

//GetIndexModel returns the index models of ClientDoc
func (c *DatatypeDoc) GetIndexModel() []mongo.IndexModel {
	return []mongo.IndexModel{{
		Keys: bsonx.Doc{
			{DatatypeDocFields.CollectionNum, bsonx.Int32(1)},
			{DatatypeDocFields.Key, bsonx.Int32(1)},
		},
	}}
}

func (c *DatatypeDoc) ToUpdateBSON() bson.D {
	return bson.D{
		{"$set", bson.D{
			{DatatypeDocFields.Key, c.Key},
			{DatatypeDocFields.CollectionNum, c.CollectionNum},
			{DatatypeDocFields.Type, c.Type},
			{DatatypeDocFields.Sseq, c.Sseq},
			{DatatypeDocFields.CreatedAt, c.CreatedAt},
		}},
		{"$currentDate", bson.D{
			{ClientDocFields.UpdatedAt, true},
		}},
	}
}
