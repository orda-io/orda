package schema

import (
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"time"
)

// CollectionDoc defines the document of Collections collection, stored in MongoDB. More specifically, it stores a number associated to the collection.
type CollectionDoc struct {
	Name      string    `bson:"_id"`
	Num       uint32    `bson:"num"`
	CreatedAt time.Time `bson:"createdAt"`
}

// CollectionDocFields defines the fields of CollectionDoc
var CollectionDocFields = struct {
	Name      string
	Num       string
	CreatedAt string
}{
	Name:      "_id",
	Num:       "num",
	CreatedAt: "createdAt",
}

// GetIndexModel returns the index models of CollectionDoc
func (c *CollectionDoc) GetIndexModel() []mongo.IndexModel {
	return []mongo.IndexModel{{
		Keys: bsonx.Doc{{Key: CollectionDocFields.Num, Value: bsonx.Int32(1)}},
	}}
}
