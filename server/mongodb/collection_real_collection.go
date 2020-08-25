package mongodb

import (
	"context"
	"github.com/knowhunger/ortoo/pkg/log"
	"github.com/knowhunger/ortoo/server/mongodb/schema"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	// ver is the field name that notes the version.
	ver = "_ver"
)

// InsertRealSnapshot inserts a snapshot for real collection.
func (r *RepositoryMongo) InsertRealSnapshot(ctx context.Context, collectionName, id string, data interface{}, sseq uint64) error {
	collection := r.db.Collection(collectionName)

	// interface{} is currently transformed to bson.M through two phases: interface{} -> bytes{} -> bson.M
	// TODO: need to develop a direct transformation method.
	marshaled, err := bson.Marshal(data)
	if err != nil {
		return log.OrtooError(err)
	}
	var bsonM = bson.M{}
	if err := bson.Unmarshal(marshaled, &bsonM); err != nil {
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
