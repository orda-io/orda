package mongodb

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"testing"
	"time"
)

func TestMongo(t *testing.T) {
	conf := &Config{
		Host:    "mongodb://root:ortoo-test@localhost:27017",
		OrtooDB: "ortoo_test",
	}
	mongo, err := New(conf)
	if err != nil {
		log.Fatal("err", err)
	}

	collection := mongo.db.Collection("hello")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	res, err := collection.InsertOne(ctx, bson.M{"name": "pi", "value": 3.14159})
	id := res.InsertedID
	fmt.Println(id)
}
