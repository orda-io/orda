package mongodb

import (
	"github.com/knowhunger/ortoo/server/mongodb/schema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const CollectionNameClient = "-_-Clients"

const PrefixSnapshot = "[SN]"

var CollectionMap = map[string]string{
	CollectionNameClient: CollectionNameClient,
}

func filterByID(ID interface{}) bson.D {
	return bson.D{bson.E{Key: schema.ID, Value: ID}}
}

func filterByName(name string) bson.D {
	return bson.D{bson.E{Key: "name", Value: name}}
}

var (
	upsert       = true
	upsertOption = &options.UpdateOptions{
		Upsert: &upsert,
	}
)
