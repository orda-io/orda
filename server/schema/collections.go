package schema

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

// CollectionDoc defines the document of Collections collection, stored in MongoDB. More specifically, it stores a number associated to the collection.
type CollectionDoc struct {
	Name      string    `bson:"_id"`
	Num       int32     `bson:"num"`
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
func (its *CollectionDoc) GetIndexModel() []mongo.IndexModel {
	return []mongo.IndexModel{{
		Keys: bsonx.Doc{{Key: CollectionDocFields.Num, Value: bsonx.Int32(1)}},
	}}
}

// GetSummary returns the summary of CollectionDoc
func (its *CollectionDoc) GetSummary() string {
	return fmt.Sprintf("%s(%d)", its.Name, its.Num)
}
