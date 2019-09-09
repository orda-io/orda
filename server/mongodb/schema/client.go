package schema

import (
	"github.com/knowhunger/ortoo/commons/model"
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

func ClientModelToBson(model *model.Client) *ClientDoc {
	return &ClientDoc{
		Cuid:       model.GetCuidString(),
		Alias:      model.Alias,
		Collection: model.Collection,
		SyncType:   model.SyncType.String(),
	}
}
