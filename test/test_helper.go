package integration

import (
	"github.com/orda-io/orda/client/pkg/context"
	"github.com/orda-io/orda/client/pkg/errors"
	"github.com/orda-io/orda/client/pkg/model"
	"github.com/orda-io/orda/client/pkg/orda"
	"github.com/orda-io/orda/server/managers"
	redis "github.com/orda-io/orda/server/redis"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/orda-io/orda/server/mongodb"
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
func GetMongo(ctx context.OrdaContext, dbName string) (*mongodb.RepositoryMongo, errors.OrdaError) {
	if m, ok := mongoDB[dbName]; ok {
		return m, nil
	}
	conf := NewTestMongoDBConfig(dbName)
	mongo, err := mongodb.New(ctx, conf)
	if err != nil {
		return nil, err
	}
	mongoDB[dbName] = mongo
	return mongo, nil
}

// NewTestOrdaClientConfig generates an OrdaClientConfig for testing.
func NewTestOrdaClientConfig(collectionName string, syncType model.SyncType) *orda.ClientConfig {
	return &orda.ClientConfig{
		ServerAddr:       "localhost:59062",
		CollectionName:   collectionName,
		NotificationAddr: "tcp://localhost:18181",
		SyncType:         syncType,
	}
}

// NewTestOrdaServerConfig generates an OrdaServerConfig for testing.
func NewTestOrdaServerConfig(dbName string) *managers.OrdaServerConfig {
	return &managers.OrdaServerConfig{
		RPCServerPort: 59062,
		RestfulPort:   59862,
		SwaggerJSON:   "../resources/orda.grpc.swagger.json",
		Notification:  "tcp://localhost:18181",
		Mongo:         NewTestMongoDBConfig(dbName),
		Redis: &redis.Config{
			Addrs: []string{"127.0.0.1:16379"},
		},
	}
}

func WaitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return true
	case <-time.After(timeout):
		return false
	}
}

// NewTestMongoDBConfig creates a new MongoDBConfig for Test
func NewTestMongoDBConfig(dbName string) *mongodb.Config {
	return &mongodb.Config{
		Host:     "localhost:27017",
		OrdaDB:   dbName,
		User:     "root",
		Password: "orda-test",
	}
}
