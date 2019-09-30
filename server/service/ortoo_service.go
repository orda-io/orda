package service

import (
	"context"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/server/mongodb"
)

//OrtooService is a rpc service of Ortoo
type OrtooService struct {
	mongo *mongodb.RepositoryMongo
}

//NewOrtooService creates a new OrtooService
func NewOrtooService(mongoConf *mongodb.Config) (*OrtooService, error) {
	mongo, err := mongodb.New(mongoConf)
	if err != nil {
		return nil, log.OrtooErrorf(err, "fail to connect to MongoDB")
	}
	return &OrtooService{
		mongo: mongo,
	}, nil
}

//Initialize initializes mongoDB and something else
func (o *OrtooService) Initialize(ctx context.Context) error {
	if err := o.mongo.InitializeCollections(ctx); err != nil {
		return log.OrtooErrorf(err, "fail to initialize mongoDB")
	}
	return nil
}
