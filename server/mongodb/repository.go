package mongodb

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

const timeout = 10 * time.Second

type mongoT struct {
	config *Config
	client *mongo.Client
	db     *mongo.Database
}

type Config struct {
	Host    string `json:"MongoHost"`
	OrtooDB string `json:"MongoOrtoo"`
}

var mongodb *mongoT

func create() (*mongoT, error) {
	//ctx, cancel := context.WithTimeout(context.Background(), timeout)
	//defer cancel()
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://root:ortoo-test@localhost:27017"))
	if err != nil {
		return nil, fmt.Errorf("fail to connect:%v", err)
	}
	//ctxPing, cancel := context.WithTimeout(context.Background(), timeout)
	//defer cancel()
	if err = client.Ping(context.TODO(), nil); err != nil {
		return nil, fmt.Errorf("fail to ping: %v", err)
	}

	db := client.Database("ortoo_test")

	return &mongoT{
		client: client,
		db:     db,
	}, nil
}

func test() {
	//options.Collection().
	//mongodb.db.CCollection()

}
