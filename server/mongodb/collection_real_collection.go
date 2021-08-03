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
	// ver is the field name that notes the version.
	ver = "_ver"
)

// InsertRealSnapshot inserts a snapshot for real collection.
func (r *RepositoryMongo) InsertRealSnapshot(
	ctx context.OrdaContext,
	collectionName string,
	id string,
	data interface{},
	sseq uint64,
) errors.OrdaError {
	collection := r.db.Collection(collectionName)

	// interface{} is currently transformed to bson.M through two phases: interface{} -> bytes{} -> bson.M
	// TODO: need to develop a direct transformation method.
	marshaled, err := bson.Marshal(data)
	if err != nil {
		return errors.ServerDBQuery.New(ctx.L(), err.Error(), data)
	}
	var bsonM = bson.M{}
	if err := bson.Unmarshal(marshaled, &bsonM); err != nil {
		return errors.ServerDBQuery.New(ctx.L(), err.Error())
	}

	bsonM[ver] = sseq
	option := &options.ReplaceOptions{}
	option.SetUpsert(true)
	res, err := collection.ReplaceOne(ctx, schema.FilterByID(id), bsonM, option)
	if err != nil {
		return errors.ServerDBQuery.New(ctx.L(), err.Error())
	}
	if res.ModifiedCount == 1 {
		ctx.L().Infof("update snapshot for key: %s in %s", id, collectionName)
	}
	return nil
}

func (r *RepositoryMongo) GetRealSnapshot(
	ctx context.OrdaContext,
	collectionName string,
	id string,
) (map[string]interface{}, errors.OrdaError) {
	collection := r.db.Collection(collectionName)
	f := schema.FilterByID(id)
	result := collection.FindOne(ctx, f)
	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, errors.ServerDBQuery.New(ctx.L(), result.Err().Error())
	}
	var snap map[string]interface{}
	if err := result.Decode(&snap); err != nil {
		return nil, errors.ServerDBDecode.New(ctx.L(), result.Err().Error())
	}
	return snap, nil
}
