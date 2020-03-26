package schema

import (
	"github.com/knowhunger/ortoo/ortoo/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"time"
)

// DatatypeDoc defines a document for datatype, stored in MongoDB
type DatatypeDoc struct {
	DUID          string    `bson:"_id"`
	Key           string    `bson:"key"`
	CollectionNum uint32    `bson:"colNum"`
	Type          string    `bson:"type"`
	SseqBegin     uint64    `bson:"sseqBegin"`
	SseqEnd       uint64    `bson:"sseqEnd"`
	Visible       bool      `bson:"visible"`
	CreatedAt     time.Time `bson:"createdAt"`
	UpdatedAt     time.Time `bson:"updatedAt"`
}

// DatatypeDocFields defines the fields of DatatypeDoc
var DatatypeDocFields = struct {
	DUID          string
	Key           string
	CollectionNum string
	Type          string
	SseqBegin     string
	SseqEnd       string
	Visible       string
	CreatedAt     string
	UpdatedAt     string
}{
	DUID:          "_id",
	Key:           "key",
	CollectionNum: "colNum",
	Type:          "type",
	SseqBegin:     "sseqBegin",
	SseqEnd:       "sseqEnd",
	Visible:       "visible",
	CreatedAt:     "createdAt",
	UpdatedAt:     "updatedAt",
}

// GetIndexModel returns the index models of the collection of ClientDoc
func (c *DatatypeDoc) GetIndexModel() []mongo.IndexModel {
	return []mongo.IndexModel{{
		Keys: bsonx.Doc{
			{DatatypeDocFields.CollectionNum, bsonx.Int32(1)},
			{DatatypeDocFields.Key, bsonx.Int32(1)},
		},
	}}
}

// ToUpdateBSON transforms DatatypeDoc to BSON type
func (c *DatatypeDoc) ToUpdateBSON() bson.D {
	return bson.D{
		{"$set", bson.D{
			{DatatypeDocFields.Key, c.Key},
			{DatatypeDocFields.CollectionNum, c.CollectionNum},
			{DatatypeDocFields.Type, c.Type},
			{DatatypeDocFields.SseqBegin, c.SseqBegin},
			{DatatypeDocFields.Visible, c.Visible},
			{DatatypeDocFields.SseqEnd, c.SseqEnd},
			{DatatypeDocFields.CreatedAt, c.CreatedAt},
		}},
		{"$currentDate", bson.D{
			{ClientDocFields.UpdatedAt, true},
		}},
	}
}

// GetType returns the type of datatype.
func (c *DatatypeDoc) GetType() model.TypeOfDatatype {
	return model.TypeOfDatatype(model.TypeOfDatatype_value[c.Type])
}
