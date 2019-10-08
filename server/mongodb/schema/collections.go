package schema

import (
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"time"
)

//CollectionDoc defines the document of Collections collection
type CollectionDoc struct {
	Name      string    `bson:"_id"`
	Num       uint32    `bson:"num"`
	CreatedAt time.Time `bson:"createdAt"`
}

//GetIndexModel returns the index models of CollectionDoc
func (c *CollectionDoc) GetIndexModel() []mongo.IndexModel {
	return []mongo.IndexModel{{
		Keys: bsonx.Doc{{Key: "num", Value: bsonx.Int32(1)}},
	}}
}
