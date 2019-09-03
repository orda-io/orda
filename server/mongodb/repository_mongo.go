package mongodb

import (
	"context"
	"fmt"
	"github.com/knowhunger/ortoo/commons/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RepositoryMongo struct {
	config           *Config
	ctx              context.Context
	client           *mongo.Client
	db               *mongo.Database
	collectionClient *CollectionClient
}

var mongodb *RepositoryMongo

func New(conf *Config) (*RepositoryMongo, error) {
	ctx := context.TODO()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(conf.Host)) //"mongodb://root:ortoo-test@localhost:27017"))
	if err != nil {
		return nil, fmt.Errorf("fail to connect:%v", err)
	}
	if err = client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("fail to ping: %v", err)
	}
	db := client.Database(conf.OrtooDB)

	return &RepositoryMongo{
		db:     db,
		ctx:    ctx,
		client: client,
	}, nil
}

func (r *RepositoryMongo) InitializeCollections() error {

	r.collectionClient = NewCollectionClient(r.ctx, r.db.Collection(CollectionNameClient))

	names, err := r.db.ListCollectionNames(r.ctx, bson.D{})
	if err != nil {
		return log.OrtooError(err, "fail to list collection names")
	}
	var realCollections = make(map[string]bool)
	for _, v := range names {
		realCollections[v] = true
	}

	if _, ok := realCollections[CollectionNameClient]; !ok {
		r.collectionClient.Create()
	}
	return nil
}
