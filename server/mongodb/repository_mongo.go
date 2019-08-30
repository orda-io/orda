package mongodb

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RepositoryMongo struct {
	config *Config
	ctx    context.Context
	client *mongo.Client
	db     *mongo.Database
}

var mongodb *RepositoryMongo

func New(ctx context.Context, conf *Config) (*RepositoryMongo, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(conf.Host)) //"mongodb://root:ortoo-test@localhost:27017"))
	if err != nil {
		return nil, fmt.Errorf("fail to connect:%v", err)
	}
	if err = client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("fail to ping: %v", err)
	}

	db := client.Database(conf.OrtooDB)
	return &RepositoryMongo{
		client: client,
		db:     db,
	}, nil
}

func (r *RepositoryMongo) InitializeCollections() {

	names, err := r.db.ListCollectionNames(r.ctx, bson.D{})
	if err != nil {

	}
	for _, v := range names {
		if collection, ok := CollectionMap[v]; !ok {

			r.db.Collection(collection)
		}
	}

}
