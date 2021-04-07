package schema

import (
	"fmt"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"strings"
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

func (its *ClientDoc) String() string {
	return fmt.Sprintf("(%d)%s:%s:%d", its.CollectionNum, its.Alias, its.CUID[0:8], len(its.CheckPoints))
}

func (its *ClientDoc) GetShortCUID() string {
	return types.ShortenUID(its.CUID)
}

// ToUpdateBSON returns a bson from a ClientDoc
func (its *ClientDoc) ToUpdateBSON() bson.D {

	var checkPointBson map[string]bson.M
	checkPointBson = make(map[string]bson.M)
	if its.CheckPoints != nil {

		for k, v := range its.CheckPoints {
			checkPointBson[k] = ToCheckPointBSON(v)
		}
	}
	return bson.D{
		{"$set", bson.D{
			{ClientDocFields.Alias, its.Alias},
			{ClientDocFields.CollectionNum, its.CollectionNum},
			{ClientDocFields.SyncType, its.SyncType},
			{ClientDocFields.CheckPoints, checkPointBson},
			{ClientDocFields.CreatedAt, its.CreatedAt},
		}},
		{"$currentDate", bson.D{
			{ClientDocFields.UpdatedAt, true},
		}},
	}

}

// GetIndexModel returns the index models of ClientDoc
func (its *ClientDoc) GetIndexModel() []mongo.IndexModel {
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
func (its *ClientDoc) GetCheckPoint(duid string) *model.CheckPoint {
	if checkPoint, ok := its.CheckPoints[duid]; ok {
		return checkPoint
	}
	return nil
}

// ClientModelToBson returns a ClientDoc from a model.Client
func ClientModelToBson(model *model.Client, collectionNum uint32) *ClientDoc {
	return &ClientDoc{
		CUID:          model.CUID,
		Alias:         model.Alias,
		CollectionNum: collectionNum,
		SyncType:      model.SyncType.String(),
	}
}

func (its *ClientDoc) GetClient() string {
	return fmt.Sprintf("%s(%s)", its.Alias, its.CUID)
}

func (its *ClientDoc) GetClientSummary() string {
	return fmt.Sprintf("%.10s(%.10s)", its.Alias, strings.ToUpper(its.CUID))
}
