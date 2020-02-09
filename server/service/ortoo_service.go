package service

import (
	"github.com/knowhunger/ortoo/server/mongodb"
	"github.com/knowhunger/ortoo/server/notification"
)

// OrtooService is a rpc service of Ortoo
type OrtooService struct {
	mongo    *mongodb.RepositoryMongo
	notifier *notification.Notifier
}

// NewOrtooService creates a new OrtooService
func NewOrtooService(mongo *mongodb.RepositoryMongo, notifier *notification.Notifier) (*OrtooService, error) {
	return &OrtooService{
		mongo:    mongo,
		notifier: notifier,
	}, nil
}

// // Initialize initializes mongoDB and something else
// func (o *OrtooService) Initialize(ctx context.Context) error {
//
// 	return nil
// }
