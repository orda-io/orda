package schema

import (
	"github.com/knowhunger/ortoo/ortoo/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"time"
)

// ClientDoc defines the document for client, stored in MongoDB.
type ClientDoc struct {
	CUID          string                       `bson:"_id"`
	Alias         string                       `bson:"alias"`
	CollectionNum uint32                       `bson:"colNum"`
	SyncType      string                       `bson:"syncType"`
	CheckPoints   map[string]*model.CheckPoint `bson:"checkpoints,omitempty"`
	CreatedAt     time.Time                    `bson:"createdAt"`
	UpdatedAt     time.Time                    `bson:"updatedAt"`
}

// ClientDocFields defines the fields of ClientDoc
var ClientDocFields = struct {
	CUID          string
	Alias         string
	CollectionNum string
	SyncType      string
	CheckPoints   string
	CreatedAt     string
	UpdatedAt     string
}{
	CUID:          "_id",
	Alias:         "alias",
	CollectionNum: "colNum",
	SyncType:      "syncType",
	CheckPoints:   "checkpoints",
	CreatedAt:     "createdAt",
	UpdatedAt:     "updatedAt",
}

// ToUpdateBSON returns a bson from a ClientDoc
func (c *ClientDoc) ToUpdateBSON() bson.D {

	var checkPointBson map[string]bson.M
	checkPointBson = make(map[string]bson.M)
	if c.CheckPoints != nil {

		for k, v := range c.CheckPoints {
			checkPointBson[k] = ToCheckPointBSON(v)
		}
	}
	return bson.D{
		{"$set", bson.D{
			{ClientDocFields.Alias, c.Alias},
			{ClientDocFields.CollectionNum, c.CollectionNum},
			{ClientDocFields.SyncType, c.SyncType},
			{ClientDocFields.CheckPoints, checkPointBson},
			{ClientDocFields.CreatedAt, c.CreatedAt},
		}},
		{"$currentDate", bson.D{
			{ClientDocFields.UpdatedAt, true},
		}},
	}

}

// GetIndexModel returns the index models of ClientDoc
func (c *ClientDoc) GetIndexModel() []mongo.IndexModel {
	return []mongo.IndexModel{
		{
			Keys: bsonx.Doc{
				{Key: ClientDocFields.CollectionNum, Value: bsonx.Int32(1)},
			},
		},
		{
			Keys: bsonx.Doc{
				{Key: ClientDocFields.CheckPoints, Value: bsonx.String("hashed")},
			},
		},
	}
}

// GetCheckPoint returns a CheckPoint of a datatype
func (c *ClientDoc) GetCheckPoint(duid string) *model.CheckPoint {
	if checkPoint, ok := c.CheckPoints[duid]; ok {
		return checkPoint
	}
	return nil
}

// ClientModelToBson returns a ClientDoc from a model.Client
func ClientModelToBson(model *model.Client, collectionNum uint32) *ClientDoc {
	return &ClientDoc{
		CUID:          model.GetCUIDString(),
		Alias:         model.Alias,
		CollectionNum: collectionNum,
		SyncType:      model.SyncType.String(),
	}
}
