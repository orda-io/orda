package integration

import (
	"context"
	"github.com/knowhunger/ortoo/ortoo"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/ortoo/model"
	"github.com/knowhunger/ortoo/server"
	"github.com/knowhunger/ortoo/server/mongodb"
	"path/filepath"
	"runtime"
	"strings"
)

var mongoDB = make(map[string]*mongodb.RepositoryMongo)

// GetFunctionName returns the function name which calls this function.
func GetFunctionName() string {
	pc, _, _, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc).Name()
	return fn[strings.LastIndex(fn, ".")+1:]
}

// GetFileName returns the file name which calls this function.
func GetFileName() string {
	_, file, _, _ := runtime.Caller(2)
	file = strings.Replace(file, ".", "_", -1)
	return filepath.Base(file)
}

// GetMongo returns an instance of RepositoryMongo for testing.
func GetMongo(dbName string) (*mongodb.RepositoryMongo, error) {
	if m, ok := mongoDB[dbName]; ok {
		return m, nil
	}
	conf := mongodb.NewTestMongoDBConfig(dbName)
	mongo, err := mongodb.New(context.Background(), conf)
	if err != nil {
		return nil, log.OrtooError(err)
	}
	mongoDB[dbName] = mongo
	return mongo, nil
}

// NewTestOrtooClientConfig generates an OrtooClientConfig for testing.
func NewTestOrtooClientConfig(collectionName string) *ortoo.ClientConfig {
	return &ortoo.ClientConfig{
		Address:          "127.0.0.1:19061",
		CollectionName:   collectionName,
		NotificationAddr: "127.0.0.1:1883",
		SyncType:         model.SyncType_NOTIFIABLE,
	}
}

// NewTestOrtooServerConfig generates an OrtooServerConfig for testing.
func NewTestOrtooServerConfig(dbName string) *server.OrtooServerConfig {
	return &server.OrtooServerConfig{
		Host:       "127.0.0.1:19061",
		PubSubAddr: "127.0.0.1:1883",
		Mongo:      mongodb.NewTestMongoDBConfig(dbName),
	}
}

// MakeTestCollection makes collections for testing.
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
	log.Logger.Infof("create a new collection:%+v", collectionDoc)
	return collectionDoc.Num, nil
}
