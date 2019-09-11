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
	*CollectionClient
	config *Config
	ctx    context.Context
	client *mongo.Client
	db     *mongo.Database
}

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

func (r *RepositoryMongo) InitializeCollections(ctx context.Context) error {

	r.CollectionClient = NewCollectionClient(r.db.Collection(CollectionNameClient))

	names, err := r.db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return log.OrtooError(err, "fail to list collection names")
	}
	var realCollections = make(map[string]bool)
	for _, v := range names {
		realCollections[v] = true
	}

	if _, ok := realCollections[CollectionNameClient]; !ok {
		if err := r.CollectionClient.Create(ctx); err != nil {
			return log.OrtooError(err, "fail to create the client collection")
		}
	}
	return nil
}

func (r *RepositoryMongo) GetOrCreateCollectionSnapshot(ctx context.Context, name string) (*CollectionSnapshot, error) {
	snapshotName := PrefixSnapshot + name
	names, err := r.db.ListCollectionNames(ctx, filterByName(snapshotName))
	if err != nil {
		return nil, log.OrtooError(err, "fail to list collections")
	}
	collection := newCollectionSnapshot(r.db.Collection(snapshotName), snapshotName)
	if len(names) == 0 {
		if err := collection.Create(ctx); err != nil {
			return nil, log.OrtooError(err, "fail to create collection")
		}
	}
	return collection, nil
}
