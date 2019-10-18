package mongodb

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//CollectionNameXXXX  is the name of the collection for XXXX
const (
	CollectionNameCounters    = "-_-Counters"
	CollectionNameClients     = "-_-Clients"
	CollectionNameCollections = "-_-Collections"
	CollectionNameDatatypes   = "-_-Datatypes"
	CollectionNameOperations  = "-_-Operations"
)

const (
	//ID is an identifier of MongoDB
	ID = "_id"
)

type filter bson.D

func (b filter) AddFilterEQ(key string, value interface{}) filter {
	return append(b, bson.E{Key: key, Value: value})
}

func (b filter) AddFilterGTE(key string, from uint32) filter {
	return append(b, bson.E{Key: key, Value: bson.D{{Key: "$gte", Value: from}}})
}

func (b filter) AddFilterLTE(key string, until uint32) filter {
	return append(b, bson.E{Key: key, Value: bson.D{{Key: "$lte", Value: until}}})
}

func GetFilter() filter {
	return filter{}
}

func filterByID(id interface{}) filter {
	return filter{bson.E{Key: ID, Value: id}}
}

func filterByName(name string) filter {
	return filter{bson.E{Key: "name", Value: name}}
}

// options
var (
	upsert       = true
	upsertOption = &options.UpdateOptions{
		Upsert: &upsert,
	}
)
