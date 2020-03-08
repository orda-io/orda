package mongodb

import (
	"context"
	"encoding/json"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/server/mongodb/schema"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	// ver is the field name that notes the version.
	ver = "_ver"
)

// InsertRealSnapshot inserts a snapshot for real collection.
func (r *RepositoryMongo) InsertRealSnapshot(ctx context.Context, collectionName, id, data string, sseq uint64) error {

	collection := r.db.Collection(collectionName)
	var bsonM = bson.M{}
	if err := json.Unmarshal([]byte(data), &bsonM); err != nil {
		return log.OrtooError(err)
	}
	bsonM[ver] = sseq
	filter := schema.GetFilter().AddSnapshot(bsonM)
	res, err := collection.UpdateOne(ctx, schema.FilterByID(id), bson.D(filter), schema.UpsertOption)
	if err != nil {
		return log.OrtooError(err)
	}
	if res.ModifiedCount == 1 {
		log.Logger.Infof("update snapshot for key: %s in %s", id, collectionName)
	}
	return nil
}
