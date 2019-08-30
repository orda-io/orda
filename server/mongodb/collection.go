package mongodb

import "go.mongodb.org/mongo-driver/mongo"

type Collection struct {
	db *mongo.Database
}

func NewCollection(db *mongo.Database) *Collection {
	return &Collection{db: db}
}
