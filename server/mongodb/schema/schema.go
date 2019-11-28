package schema

import (
	"github.com/knowhunger/ortoo/commons/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CollectionNameXXXX  is the name of the collection for XXXX
const (
	CollectionNameCounters    = "-_-Counters"
	CollectionNameClients     = "-_-Clients"
	CollectionNameCollections = "-_-Collections"
	CollectionNameDatatypes   = "-_-Datatypes"
	CollectionNameOperations  = "-_-Operations"
	CollectionNameSnapshot    = "-_-Snapshots"
)

const (
	// ID is an identifier of MongoDB
	ID = "_id"
)

type filter bson.D

func (b filter) AddFilterEQ(key string, value interface{}) filter {
	return append(b, bson.E{Key: key, Value: value})
}

func (b filter) AddFilterGTE(key string, from interface{}) filter {
	return append(b, bson.E{Key: key, Value: bson.D{{Key: "$gte", Value: from}}})
}

func (b filter) AddFilterLTE(key string, to interface{}) filter {
	return append(b, bson.E{Key: key, Value: bson.D{{Key: "$lte", Value: to}}})
}

func ToCheckPointBSON(checkPoint *model.CheckPoint) bson.M {
	return bson.M{"s": checkPoint.Sseq, "c": checkPoint.Cseq}
}

func GetFilter() filter {
	return filter{}
}

func FilterByID(id interface{}) filter {
	return filter{bson.E{Key: ID, Value: id}}
}

func FilterByName(name string) filter {
	return filter{bson.E{Key: "name", Value: name}}
}

// options
var (
	upsert       = true
	UpsertOption = &options.UpdateOptions{
		Upsert: &upsert,
	}
)
