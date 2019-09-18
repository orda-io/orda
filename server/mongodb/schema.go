package mongodb

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	//CollectionNameClients is the name of the collection for Clients
	CollectionNameClients = "-_-Clients"
	//CollectionNameCollections is the name of the collection for Collections
	CollectionNameCollections = "-_-Collections"
)

//var CollectionMap = map[string]string{
//	CollectionNameClients:     CollectionNameClients,
//	CollectionNameCollections: CollectionNameCollections,
//}

const (
	//ID is an identifier of MongoDB
	ID = "_id"
)

func filterByID(id interface{}) bson.D {
	return bson.D{bson.E{Key: ID, Value: id}}
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
