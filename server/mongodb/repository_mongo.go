package mongodb

import (
	"context"
	"github.com/knowhunger/ortoo/pkg/log"
	"github.com/knowhunger/ortoo/server/mongodb/schema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// RepositoryMongo is a tool struct for MongoDB
type RepositoryMongo struct {
	*MongoCollections
	config *Config
	ctx    context.Context
	client *mongo.Client
	db     *mongo.Database
}

// New creates a new RepositoryMongo
func New(ctx context.Context, conf *Config) (*RepositoryMongo, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(conf.Host)) // "mongodb://root:ortoo-test@localhost:27017"))
	if err != nil {
		return nil, log.OrtooErrorf(err, "fail to connect MongoDB")
	}
	if err = client.Ping(ctx, nil); err != nil {
		return nil, log.OrtooErrorf(err, "fail to ping MongoDB")
	}
	db := client.Database(conf.OrtooDB)
	log.Logger.Infof("New MongoDB:%v", conf.OrtooDB)
	repo := &RepositoryMongo{
		db:     db,
		ctx:    ctx,
		client: client,
		MongoCollections: &MongoCollections{
			mongoClient: client,
		},
	}
	if err := repo.InitializeCollections(ctx); err != nil {
		return nil, log.OrtooError(err)
	}

	return repo, nil
}

// InitializeCollections initialize collections
func (r *RepositoryMongo) InitializeCollections(ctx context.Context) error {

	r.clients = r.db.Collection(schema.CollectionNameClients)
	r.counters = r.db.Collection(schema.CollectionNameColNumGenerator)
	r.snapshots = r.db.Collection(schema.CollectionNameSnapshot)
	r.datatypes = r.db.Collection(schema.CollectionNameDatatypes)
	r.operations = r.db.Collection(schema.CollectionNameOperations)
	r.collections = r.db.Collection(schema.CollectionNameCollections)

	names, err := r.db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return log.OrtooErrorf(err, "fail to list collection names")
	}
	var realCollections = make(map[string]bool)
	for _, v := range names {
		realCollections[v] = true
	}

	if _, ok := realCollections[schema.CollectionNameClients]; !ok {
		if err := r.create(ctx, r.clients, &schema.ClientDoc{}); err != nil {
			return log.OrtooErrorf(err, "fail to create the client collection")
		}
	}
	if _, ok := realCollections[schema.CollectionNameDatatypes]; !ok {
		if err := r.create(ctx, r.datatypes, &schema.DatatypeDoc{}); err != nil {
			return log.OrtooErrorf(err, "fail to create the collections collection")
		}
	}
	if _, ok := realCollections[schema.CollectionNameOperations]; !ok {
		if err := r.create(ctx, r.operations, &schema.OperationDoc{}); err != nil {
			return log.OrtooErrorf(err, "fail to create the operations collection")
		}
	}
	if _, ok := realCollections[schema.CollectionNameSnapshot]; !ok {
		if err := r.create(ctx, r.snapshots, &schema.SnapshotDoc{}); err != nil {
			return log.OrtooErrorf(err, "fail to create snapshots collection")
		}
	}
	if _, ok := realCollections[schema.CollectionNameCollections]; !ok {
		if err := r.create(ctx, r.collections, &schema.CollectionDoc{}); err != nil {
			return log.OrtooErrorf(err, "fail to create the collections collection")
		}
	}
	return nil
}

// ResetCollections resets all collections related to collectionName
func (r *RepositoryMongo) ResetCollections(ctx context.Context, collectionName string) error {
	if err := r.PurgeAllDocumentsOfCollection(ctx, collectionName); err != nil {
		return log.OrtooError(err)
	}
	collection := r.db.Collection(collectionName)
	if err := collection.Drop(ctx); err != nil {
		return log.OrtooError(err)
	}
	return nil
}

// GetOrCreateRealCollection is a method that gets or creates a collection of snapshot
func (r *RepositoryMongo) GetOrCreateRealCollection(ctx context.Context, name string) error {

	names, err := r.db.ListCollectionNames(ctx, schema.FilterByName(name))
	if err != nil {
		return log.OrtooErrorf(err, "fail to list collections")
	}
	collection := r.db.Collection(name)
	if len(names) == 0 {
		if err := r.create(ctx, collection, nil); err != nil {
			return log.OrtooErrorf(err, "fail to create collection")
		}
	}
	return nil
}

// MakeCollection makes a real collection.
func MakeCollection(mongo *RepositoryMongo, collectionName string) (uint32, error) {
	collectionDoc, err := mongo.GetCollection(context.TODO(), collectionName)
	if err != nil {
		return 0, log.OrtooError(err)
	}
	if collectionDoc != nil {
		return collectionDoc.Num, nil
	}
	collectionDoc, err = mongo.InsertCollection(context.TODO(), collectionName)
	if err != nil {
		return 0, log.OrtooError(err)
	}
	log.Logger.Infof("create a new collection:%+v", collectionDoc)
	return collectionDoc.Num, nil
}
