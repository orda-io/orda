package test_helper

import (
	"context"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/server/mongodb"
	"path/filepath"
	"runtime"
	"strings"
)

var mongoDB = make(map[string]*mongodb.RepositoryMongo)

func GetFunctionName() string {
	pc, _, _, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc).Name()
	return fn[strings.LastIndex(fn, ".")+1:]
}

func GetFileName() string {
	_, file, _, _ := runtime.Caller(2)
	file = strings.Replace(file, ".", "_", -1)
	return filepath.Base(file)
}

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
