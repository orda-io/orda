package integration

import (
	"context"
	"github.com/knowhunger/ortoo/commons"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/server"
	"github.com/knowhunger/ortoo/server/mongodb"
)

// NewTestOrtooClientConfig ...
func NewTestOrtooClientConfig(collectionName string) *commons.OrtooClientConfig {
	return &commons.OrtooClientConfig{
		Address:        "127.0.0.1",
		Port:           19061,
		CollectionName: collectionName,
	}
}

// NewTestOrtooServerConfig ...
func NewTestOrtooServerConfig(dbName string) *server.OrtooServerConfig {
	return &server.OrtooServerConfig{
		Host:  "127.0.0.1",
		Port:  19061,
		Mongo: mongodb.NewTestMongoDBConfig(dbName),
	}
}

func MakeTestCollection(mongo *mongodb.RepositoryMongo, collectionName string) (uint32, error) {
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
	log.Logger.Infof("a new collection is created:%+v", collectionDoc)
	return collectionDoc.Num, nil
}
