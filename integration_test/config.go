package integration

import (
	"context"
	"github.com/knowhunger/ortoo/commons"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/server"
	"github.com/knowhunger/ortoo/server/mongodb"
)

//NewTestOrtooClientConfig ...
func NewTestOrtooClientConfig(dbName string, collectionName string) *commons.OrtooClientConfig {
	return &commons.OrtooClientConfig{
		Address:        "127.0.0.1",
		Port:           19061,
		CollectionName: collectionName,
		Alias:          dbName,
	}
}

//NewTestMongoDBConfig ...
func NewTestMongoDBConfig(dbName string) *mongodb.Config {
	return &mongodb.Config{
		Host:    "mongodb://root:ortoo-test@localhost:27017",
		OrtooDB: dbName,
	}
}

//NewTestOrtooServerConfig ...
func NewTestOrtooServerConfig(dbName string) *server.OrtooServerConfig {
	return &server.OrtooServerConfig{
		Host:  "127.0.0.1",
		Port:  19061,
		Mongo: NewTestMongoDBConfig(dbName),
	}
}

func MakeTestCollection(mongo *mongodb.RepositoryMongo, collectionName string) error {
	collectionDoc, err := mongo.GetCollection(context.TODO(), collectionName)
	if err != nil {
		return log.OrtooError(err)
	}
	if collectionDoc != nil {
		return nil
	}
	collectionDoc, err = mongo.InsertCollection(context.TODO(), collectionName)
	if err != nil {
		return log.OrtooError(err)
	}
	log.Logger.Infof("a new collection is created:%s", collectionDoc)
	return nil
}
