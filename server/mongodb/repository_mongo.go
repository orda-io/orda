package mongodb

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/orda-io/orda/client/pkg/errors"
	"github.com/orda-io/orda/client/pkg/iface"
	"github.com/orda-io/orda/server/schema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io/ioutil"
)

// RepositoryMongo is a tool struct for MongoDB
type RepositoryMongo struct {
	*MongoCollections
	config *Config
	client *mongo.Client
	db     *mongo.Database
}

// New creates a new RepositoryMongo
func New(ctx iface.OrdaContext, conf *Config) (*RepositoryMongo, errors.OrdaError) {

	option := options.Client().ApplyURI(conf.getConnectionString())
	if conf.CertFile != "" {
		tlsConfig, err := getCustomTLSConfig(ctx, conf.CertFile)
		if err != nil {
			return nil, err
		}
		option.SetTLSConfig(tlsConfig)
	}

	client, err := mongo.Connect(ctx, option)
	if err != nil {
		return nil, errors.ServerDBQuery.New(ctx.L(), err.Error())
	}
	if err = client.Ping(ctx, nil); err != nil {
		return nil, errors.ServerDBQuery.New(ctx.L(), err.Error())
	}
	db := client.Database(conf.OrdaDB)
	ctx.L().Infof("New MongoDB:%v", conf.OrdaDB)
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

func getCustomTLSConfig(ctx iface.OrdaContext, caFile string) (*tls.Config, errors.OrdaError) {
	tlsConfig := new(tls.Config)
	certs, err := ioutil.ReadFile(caFile)

	if err != nil {
		return tlsConfig, errors.ServerDBInit.New(ctx.L(), err.Error())
	}
	tlsConfig.RootCAs = x509.NewCertPool()
	if ok := tlsConfig.RootCAs.AppendCertsFromPEM(certs); !ok {
		return tlsConfig, errors.ServerDBInit.New(ctx.L(), err.Error())
	}
	return tlsConfig, nil
}

// InitializeCollections initializes collections
func (its *RepositoryMongo) InitializeCollections(ctx iface.OrdaContext) errors.OrdaError {
	its.clients = its.db.Collection(schema.CollectionNameClients)
	its.counters = its.db.Collection(schema.CollectionNameColNumGenerator)
	its.snapshots = its.db.Collection(schema.CollectionNameSnapshot)

	its.datatypes = its.db.Collection(schema.CollectionNameDatatypes)
	its.operations = its.db.Collection(schema.CollectionNameOperations)
	its.collections = its.db.Collection(schema.CollectionNameCollections)

	names, err := its.db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return errors.ServerDBQuery.New(ctx.L(), err.Error())
	}
	var realCollections = make(map[string]bool)
	for _, v := range names {
		realCollections[v] = true
	}

	if _, ok := realCollections[schema.CollectionNameClients]; !ok {
		if err := its.createCollection(ctx, its.clients, &schema.ClientDoc{}); err != nil {
			return errors.ServerDBQuery.New(ctx.L(), schema.CollectionNameClients+err.Error())
		}
	}
	if _, ok := realCollections[schema.CollectionNameDatatypes]; !ok {
		if err := its.createCollection(ctx, its.datatypes, &schema.DatatypeDoc{}); err != nil {
			return errors.ServerDBQuery.New(ctx.L(), schema.CollectionNameDatatypes+err.Error())
		}
	}
	if _, ok := realCollections[schema.CollectionNameOperations]; !ok {
		if err := its.createCollection(ctx, its.operations, &schema.OperationDoc{}); err != nil {
			return errors.ServerDBQuery.New(ctx.L(), schema.CollectionNameOperations+err.Error())
		}
	}
	if _, ok := realCollections[schema.CollectionNameSnapshot]; !ok {
		if err := its.createCollection(ctx, its.snapshots, &schema.SnapshotDoc{}); err != nil {
			return errors.ServerDBQuery.New(ctx.L(), schema.CollectionNameSnapshot+err.Error())
		}
	}
	if _, ok := realCollections[schema.CollectionNameCollections]; !ok {
		if err := its.createCollection(ctx, its.collections, &schema.CollectionDoc{}); err != nil {
			return errors.ServerDBQuery.New(ctx.L(), schema.CollectionNameCollections+err.Error())
		}
	}
	return nil
}

// PurgeCollection purges all collections related to collectionName
func (its *RepositoryMongo) PurgeCollection(ctx iface.OrdaContext, collectionName string) errors.OrdaError {
	if err := its.PurgeAllDocumentsOfCollection(ctx, collectionName); err != nil {
		return errors.ServerDBQuery.New(ctx.L(), err.Error())
	}

	collection := its.db.Collection(collectionName)
	if err := collection.Drop(ctx); err != nil {
		return errors.ServerDBQuery.New(ctx.L(), err.Error())
	}
	return nil
}

// GetOrCreateRealCollection is a method that gets or creates a collection of snapshot
func (its *RepositoryMongo) GetOrCreateRealCollection(ctx iface.OrdaContext, name string) errors.OrdaError {
	names, err := its.db.ListCollectionNames(ctx, schema.FilterByName(name))
	if err != nil {
		return errors.ServerDBQuery.New(ctx.L(), err.Error())
	}

	if len(names) == 0 {
		return its.createCollection(ctx, its.db.Collection(name), nil)
	}
	return nil
}

// Close closes the repository of MongoDB
func (its *RepositoryMongo) Close(ctx iface.OrdaContext) errors.OrdaError {
	if err := its.mongoClient.Disconnect(ctx); err != nil {
		return errors.ServerDBQuery.New(ctx.L(), err.Error())
	}
	return nil
}

// MakeCollection makes a real collection.
func MakeCollection(ctx iface.OrdaContext, mongo *RepositoryMongo, collectionName string) (int32, errors.OrdaError) {
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
