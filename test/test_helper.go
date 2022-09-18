package integration

import (
	gocontext "context"
	"github.com/orda-io/orda/client/pkg/context"
	"github.com/orda-io/orda/client/pkg/errors"
	"github.com/orda-io/orda/client/pkg/iface"
	"github.com/orda-io/orda/client/pkg/model"
	"github.com/orda-io/orda/client/pkg/orda"
	"github.com/orda-io/orda/server/constants"
	"github.com/orda-io/orda/server/managers"
	redis "github.com/orda-io/orda/server/redis"
	"github.com/stretchr/testify/require"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
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
func GetMongo(ctx iface.OrdaContext, dbName string) (*mongodb.RepositoryMongo, errors.OrdaError) {
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

// WaitTimeout waits for timeout of the WaitGroup during the specified duration
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

// NewTestManagers creates a new Managers for Test
func NewTestManagers(ctx iface.OrdaContext, dbName string) (*managers.Managers, errors.OrdaError) {
	conf := NewTestOrdaServerConfig(dbName)
	return managers.New(ctx, conf)
}

// InitTestDBCollection initializes db collection for testing
func InitTestDBCollection(t *testing.T, dbName string) (*mongodb.RepositoryMongo, iface.OrdaContext, int32) {
	ctx := context.NewOrdaContext(gocontext.TODO(), constants.TagTest).
		UpdateCollectionTags(t.Name(), 0)
	mongo, err := GetMongo(ctx, dbName)
	require.NoError(t, err)

	err = mongo.PurgeCollection(ctx, t.Name())
	require.NoError(t, err)
	collectionNum, err := mongodb.MakeCollection(ctx, mongo, t.Name())
	require.NoError(t, err)
	ctx.UpdateCollectionTags(t.Name(), collectionNum)
	ctx.L().Infof("Init Test DB Collection %v(%d) in %v", t.Name(), collectionNum, dbName)
	return mongo, ctx, collectionNum
}
