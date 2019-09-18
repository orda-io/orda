package schema

import (
	"github.com/knowhunger/ortoo/commons/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"time"
)

//ClientDoc defines the document related to client
type ClientDoc struct {
	Cuid       string    `bson:"_id"`
	Alias      string    `bson:"alias"`
	Collection string    `bson:"collection"`
	SyncType   string    `bson:"syncType"`
	CreatedAt  time.Time `bson:"createdAt"`
	UpdatedAt  time.Time `bson:"updatedAt"`
}

//ToUpdateBson returns a bson from a ClientDoc
func (c *ClientDoc) ToUpdateBson() bson.D {
	return bson.D{
		{"$set", bson.D{
			{"alias", c.Alias},
			{"collection", c.Collection},
			{"createdAt", c.CreatedAt},
			{"syncType", c.SyncType},
		}},
		{"$currentDate", bson.D{
			{"updatedAt", true},
		}},
	}
}

//GetIndexModel returns the index models of ClientDoc
func (c *ClientDoc) GetIndexModel() []mongo.IndexModel {
	return []mongo.IndexModel{{
		Keys: bsonx.Doc{{Key: "collection", Value: bsonx.Int32(1)}},
	}}
}

//ClientModelToBson returns a ClientDoc from a model.Client
func ClientModelToBson(model *model.Client) *ClientDoc {
	return &ClientDoc{
		Cuid:       model.GetCuidString(),
		Alias:      model.Alias,
		Collection: model.Collection,
		SyncType:   model.SyncType.String(),
	}
}
