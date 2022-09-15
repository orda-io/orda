package schema

import (
	"fmt"
	"github.com/orda-io/orda/client/pkg/model"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

// ClientDoc defines the document for client, stored in MongoDB.
type ClientDoc struct {
	CUID          string    `bson:"_id"`
	Alias         string    `bson:"alias"`
	CollectionNum int32     `bson:"colNum"`
	Type          int8      `bson:"type"`
	SyncType      int8      `bson:"syncType"`
	CreatedAt     time.Time `bson:"createdAt"`
	UpdatedAt     time.Time `bson:"updatedAt"`
}

// ClientDocFields defines the fields of ClientDoc
var ClientDocFields = struct {
	CUID          string
	Alias         string
	CollectionNum string
	Type          string
	SyncType      string
	CreatedAt     string
	UpdatedAt     string
}{
	CUID:          "_id",
	Alias:         "alias",
	CollectionNum: "colNum",
	Type:          "type",
	SyncType:      "syncType",
	CreatedAt:     "createdAt",
	UpdatedAt:     "updatedAt",
}

func (its *ClientDoc) String() string {
	return fmt.Sprintf("(%d)%s:%s", its.CollectionNum, its.Alias, its.CUID)
}

// ToUpdateBSON returns a bson from a ClientDoc
func (its *ClientDoc) ToUpdateBSON() bson.D {
	return bson.D{
		{"$set", bson.D{
			{ClientDocFields.Alias, its.Alias},
			{ClientDocFields.CollectionNum, its.CollectionNum},
			{ClientDocFields.Type, its.Type},
			{ClientDocFields.SyncType, its.SyncType},
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
				{Key: ClientDocFields.Type, Value: bsonx.Int32(1)},
			},
		},
	}
}

// ClientModelToBson returns a ClientDoc from a model.Client
func ClientModelToBson(model *model.Client, collectionNum int32) *ClientDoc {
	return &ClientDoc{
		CUID:          model.CUID,
		Alias:         model.Alias,
		CollectionNum: collectionNum,
		Type:          int8(model.Type),
		SyncType:      int8(model.SyncType),
	}
}

// GetModel returns model.Client
func (its *ClientDoc) GetModel() *model.Client {
	return &model.Client{
		CUID:       its.CUID,
		Alias:      its.Alias,
		Collection: strconv.Itoa(int(its.CollectionNum)),
		Type:       model.ClientType(its.Type),
		SyncType:   model.SyncType(its.SyncType),
	}
}

// ToString returns ClientDoc string
func (its *ClientDoc) ToString() string {
	return fmt.Sprintf("%s(%s)", its.Alias, its.CUID)
}

// GetType returns model.ClientType
func (its *ClientDoc) GetType() model.ClientType {
	return model.ClientType(its.Type)
}
