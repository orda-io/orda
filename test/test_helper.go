package integration

import (
	"github.com/knowhunger/ortoo/pkg/context"
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/pkg/ortoo"
	"github.com/knowhunger/ortoo/server/mongodb"
	"github.com/knowhunger/ortoo/server/server"
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
func GetMongo(ctx context.OrtooContext, dbName string) (*mongodb.RepositoryMongo, errors.OrtooError) {
	if m, ok := mongoDB[dbName]; ok {
		return m, nil
	}
	conf := mongodb.NewTestMongoDBConfig(dbName)
	mongo, err := mongodb.New(ctx, conf)
	if err != nil {
		return nil, err
	}
	mongoDB[dbName] = mongo
	return mongo, nil
}

// NewTestOrtooClientConfig generates an OrtooClientConfig for testing.
func NewTestOrtooClientConfig(collectionName string) *ortoo.ClientConfig {
	return &ortoo.ClientConfig{
		ServerAddr:       "127.0.0.1:19061",
		CollectionName:   collectionName,
		NotificationAddr: "127.0.0.1:11883",
		SyncType:         model.SyncType_NOTIFIABLE,
	}
}

// NewTestOrtooServerConfig generates an OrtooServerConfig for testing.
func NewTestOrtooServerConfig(dbName string) *server.OrtooServerConfig {
	return &server.OrtooServerConfig{
		RPCServerPort: 19061,
		RestfulPort:   19861,
		Notification:  "127.0.0.1:11883",
		Mongo:         *mongodb.NewTestMongoDBConfig(dbName),
	}
}
