package mongodb

import (
	"github.com/knowhunger/ortoo/pkg/context"
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/server/mongodb/schema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// RepositoryMongo is a tool struct for MongoDB
type RepositoryMongo struct {
	*MongoCollections
	config *Config
	client *mongo.Client
	db     *mongo.Database
}

// New creates a new RepositoryMongo
func New(ctx context.OrtooContext, conf *Config) (*RepositoryMongo, errors.OrtooError) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(conf.Host)) // "mongodb://root:ortoo-test@localhost:27017"))
	if err != nil {
		return nil, errors.ServerDBQuery.New(ctx.L(), err.Error())
	}
	if err = client.Ping(ctx, nil); err != nil {
		return nil, errors.ServerDBQuery.New(ctx.L(), err.Error())
	}
	db := client.Database(conf.OrtooDB)
	ctx.L().Infof("New MongoDB:%v", conf.OrtooDB)
	repo := &RepositoryMongo{
		db:     db,
		client: client,
		MongoCollections: &MongoCollections{
			mongoClient: client,
		},
	}
	if err := repo.InitializeCollections(ctx); err != nil {
		return nil, err
	}

	return repo, nil
}

// InitializeCollections initializes collections
func (r *RepositoryMongo) InitializeCollections(ctx context.OrtooContext) errors.OrtooError {

	r.clients = r.db.Collection(schema.CollectionNameClients)
	r.counters = r.db.Collection(schema.CollectionNameColNumGenerator)
	r.snapshots = r.db.Collection(schema.CollectionNameSnapshot)
	r.datatypes = r.db.Collection(schema.CollectionNameDatatypes)
	r.operations = r.db.Collection(schema.CollectionNameOperations)
	r.collections = r.db.Collection(schema.CollectionNameCollections)

	names, err := r.db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return errors.ServerDBQuery.New(ctx.L(), err.Error())
	}
	var realCollections = make(map[string]bool)
	for _, v := range names {
		realCollections[v] = true
	}

	if _, ok := realCollections[schema.CollectionNameClients]; !ok {
		if err := r.createCollection(ctx, r.clients, &schema.ClientDoc{}); err != nil {
			return errors.ServerDBQuery.New(ctx.L(), schema.CollectionNameClients+err.Error())
		}
	}
	if _, ok := realCollections[schema.CollectionNameDatatypes]; !ok {
		if err := r.createCollection(ctx, r.datatypes, &schema.DatatypeDoc{}); err != nil {
			return errors.ServerDBQuery.New(ctx.L(), schema.CollectionNameDatatypes+err.Error())
		}
	}
	if _, ok := realCollections[schema.CollectionNameOperations]; !ok {
		if err := r.createCollection(ctx, r.operations, &schema.OperationDoc{}); err != nil {
			return errors.ServerDBQuery.New(ctx.L(), schema.CollectionNameOperations+err.Error())
		}
	}
	if _, ok := realCollections[schema.CollectionNameSnapshot]; !ok {
		if err := r.createCollection(ctx, r.snapshots, &schema.SnapshotDoc{}); err != nil {
			return errors.ServerDBQuery.New(ctx.L(), schema.CollectionNameSnapshot+err.Error())
		}
	}
	if _, ok := realCollections[schema.CollectionNameCollections]; !ok {
		if err := r.createCollection(ctx, r.collections, &schema.CollectionDoc{}); err != nil {
			return errors.ServerDBQuery.New(ctx.L(), schema.CollectionNameCollections+err.Error())
		}
	}
	return nil
}

// ResetCollections resets all collections related to collectionName
func (r *RepositoryMongo) ResetCollections(ctx context.OrtooContext, collectionName string) errors.OrtooError {
	if err := r.PurgeAllDocumentsOfCollection(ctx, collectionName); err != nil {
		return errors.ServerDBQuery.New(ctx.L(), err.Error())
	}
	collection := r.db.Collection(collectionName)
	if err := collection.Drop(ctx); err != nil {
		return errors.ServerDBQuery.New(ctx.L(), err.Error())
	}
	return nil
}

// GetOrCreateRealCollection is a method that gets or creates a collection of snapshot
func (r *RepositoryMongo) GetOrCreateRealCollection(ctx context.OrtooContext, name string) errors.OrtooError {

	names, err := r.db.ListCollectionNames(ctx, schema.FilterByName(name))
	if err != nil {
		return errors.ServerDBQuery.New(ctx.L(), err.Error())
	}

	if len(names) == 0 {
		return r.createCollection(ctx, r.db.Collection(name), nil)
	}
	return nil
}

// MakeCollection makes a real collection.
func MakeCollection(ctx context.OrtooContext, mongo *RepositoryMongo, collectionName string) (uint32, errors.OrtooError) {
	collectionDoc, err := mongo.GetCollection(ctx, collectionName)
	if err != nil {
		return 0, err
	}
	if collectionDoc != nil {
		return collectionDoc.Num, nil
	}
	collectionDoc, err = mongo.InsertCollection(ctx, collectionName)
	if err != nil {
		return 0, err
	}
	ctx.L().Infof("create a new collection:%+v", collectionDoc)
	return collectionDoc.Num, nil
}
