package mongodb

import (
	"github.com/orda-io/orda/pkg/context"
	"github.com/orda-io/orda/pkg/errors"
	"github.com/orda-io/orda/server/mongodb/schema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	idForCollection = "collectionID"
)

// GetNextCollectionNum gets a collection number that is assigned to a collection.
func (its *MongoCollections) GetNextCollectionNum(ctx context.OrdaContext) (uint32, errors.OrdaError) {
	opts := options.FindOneAndUpdate()
	opts.SetUpsert(true)
	var update = bson.M{
		"$inc": bson.M{schema.CounterDocFields.Num: 1},
	}

	result := its.counters.FindOneAndUpdate(ctx, schema.FilterByID(idForCollection), update, opts)
	if err := result.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return 1, nil
		}
		return 0, errors.ServerDBQuery.New(ctx.L(), err.Error())
	}
	var counterDoc = schema.CounterDoc{}
	if err := result.Decode(&counterDoc); err != nil {
		return 0, errors.ServerDBDecode.New(ctx.L(), err.Error())
	}
	return counterDoc.Num, nil
}
