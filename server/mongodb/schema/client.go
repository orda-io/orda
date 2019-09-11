package schema

import (
	"github.com/knowhunger/ortoo/commons/model"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

type ClientDoc struct {
	Cuid       string    `bson:"_id"`
	Alias      string    `bson:"alias"`
	Collection string    `bson:"collection"`
	SyncType   string    `bson:"syncType"`
	CreatedAt  time.Time `bson:"createdAt"`
	UpdatedAt  time.Time `bson:"updatedAt"`
}

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

func ClientModelToBson(model *model.Client) *ClientDoc {
	return &ClientDoc{
		Cuid:       model.GetCuidString(),
		Alias:      model.Alias,
		Collection: model.Collection,
		SyncType:   model.SyncType.String(),
	}
}
