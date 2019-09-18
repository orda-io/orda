package mongodb

import (
	"context"
	"fmt"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/server/mongodb/schema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//RepositoryMongo is a tool struct for MongoDB
type RepositoryMongo struct {
	*CollectionClients
	*CollectionCollections
	config *Config
	ctx    context.Context
	client *mongo.Client
	db     *mongo.Database
}

//New creates a new RepositoryMongo
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

//InitializeCollections initialize collections
func (r *RepositoryMongo) InitializeCollections(ctx context.Context) error {

	r.CollectionClients = NewCollectionClients(r.db.Collection(CollectionNameClients))
	r.CollectionCollections = NewCollectionCollections(r.db.Collection(CollectionNameCollections))

	names, err := r.db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return log.OrtooError(err, "fail to list collection names")
	}
	var realCollections = make(map[string]bool)
	for _, v := range names {
		realCollections[v] = true
	}

	if _, ok := realCollections[CollectionNameClients]; !ok {
		if err := r.CollectionClients.create(ctx, &schema.ClientDoc{}); err != nil {
			return log.OrtooError(err, "fail to create the client collection")
		}
	}
	if _, ok := realCollections[CollectionNameCollections]; !ok {
		if err := r.CollectionCollections.create(ctx, &schema.CollectionDoc{}); err != nil {
			return log.OrtooError(err, "fail to create the collections collection")
		}
	}
	return nil
}

//GetOrCreateCollectionSnapshot is a method that gets or creates a collection of snapshot
func (r *RepositoryMongo) GetOrCreateCollectionSnapshot(ctx context.Context, name string) (*CollectionSnapshots, error) {

	names, err := r.db.ListCollectionNames(ctx, filterByName(name))
	if err != nil {
		return nil, log.OrtooError(err, "fail to list collections")
	}
	collection := newCollectionSnapshot(r.db.Collection(name), name)
	if len(names) == 0 {
		if err := collection.create(ctx, nil); err != nil {
			return nil, log.OrtooError(err, "fail to create collection")
		}
	}
	return collection, nil
}
