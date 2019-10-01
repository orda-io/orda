package schema

import (
	"go.mongodb.org/mongo-driver/mongo"
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
	return []mongo.IndexModel{{}}
}
