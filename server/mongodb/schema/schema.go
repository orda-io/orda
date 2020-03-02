package schema

import (
	"fmt"
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

// Filter is the type used to filter
type Filter bson.D

// AddFilterEQ is a function to add EQ to Filter
func (b Filter) AddFilterEQ(key string, value interface{}) Filter {
	return append(b, bson.E{Key: key, Value: value})
}

// AddFilterGTE is a function to add GTE to Filter
func (b Filter) AddFilterGTE(key string, from interface{}) Filter {
	return append(b, bson.E{Key: key, Value: bson.D{{Key: "$gte", Value: from}}})
}

// AddFilterLTE is a function to add LTE to Filter
func (b Filter) AddFilterLTE(key string, to interface{}) Filter {
	return append(b, bson.E{Key: key, Value: bson.D{{Key: "$lte", Value: to}}})
}

// ToCheckPointBSON is a function to change a checkpoint to BSON
func ToCheckPointBSON(checkPoint *model.CheckPoint) bson.M {
	return bson.M{"s": checkPoint.Sseq, "c": checkPoint.Cseq}
}

// GetFilter returns an instance of Filter
func GetFilter() Filter {
	return Filter{}
}

// FilterByID returns an instance of Filter of ID
func FilterByID(id interface{}) Filter {
	return Filter{bson.E{Key: ID, Value: id}}
}

// FilterByName returns an instance of Filter of Name
func FilterByName(name string) Filter {
	return Filter{bson.E{Key: "name", Value: name}}
}

// AddSnapshot adds a snapshot value
func (b Filter) AddSnapshot(data bson.M) Filter {
	return append(b, bson.E{Key: "$set", Value: data})
}

// AddSetCheckPoint adds the Filter that updates checkpoint
func (b Filter) AddSetCheckPoint(key string, checkPoint *model.CheckPoint) Filter {
	return append(b, bson.E{Key: "$set", Value: bson.D{
		{Key: fmt.Sprintf("%s.%s", ClientDocFields.CheckPoints, key), Value: ToCheckPointBSON(checkPoint)},
	}})
}

// AddUnsetCheckPoint adds the Filter that unsets the checkpoint
func (b Filter) AddUnsetCheckPoint(key string) Filter {
	return append(b, bson.E{Key: "$unset", Value: bson.D{
		{Key: fmt.Sprintf("%s.%s", ClientDocFields.CheckPoints, key), Value: 1},
	}})
}

// AddExists adds the Filter which examines the existence of the key
func (b Filter) AddExists(key string) Filter {
	return append(b, bson.E{Key: key, Value: bson.D{
		{Key: "$exists", Value: true},
	}})
}

// options
var (
	upsert       = true
	UpsertOption = &options.UpdateOptions{
		Upsert: &upsert,
	}
)
