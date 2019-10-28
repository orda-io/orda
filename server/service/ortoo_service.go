package service

import (
	"context"
	"github.com/knowhunger/ortoo/server/mongodb"
)

//OrtooService is a rpc service of Ortoo
type OrtooService struct {
	mongo *mongodb.RepositoryMongo
}

//NewOrtooService creates a new OrtooService
func NewOrtooService(mongo *mongodb.RepositoryMongo) (*OrtooService, error) {
	return &OrtooService{
		mongo: mongo,
	}, nil
}

//Initialize initializes mongoDB and something else
func (o *OrtooService) Initialize(ctx context.Context) error {

	return nil
}
