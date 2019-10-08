package integration

import (
	"context"
	"fmt"
	"github.com/knowhunger/ortoo/server/mongodb"
	"log"
	"testing"
)

func TestMongo(t *testing.T) {
	conf := NewTestMongoDBConfig("ortoo_unit_test")
	mongo, err := mongodb.New(conf)
	if err != nil {
		log.Fatalf("fail to create mongoDB instance:%v", err)
	}
	if err := mongo.InitializeCollections(context.TODO()); err != nil {
		log.Fatalf("fail to initialize collection:%v", err)
	}
	//wg := sync.WaitGroup{}
	//wg.Add(10)
	//for i:=0;i<10;i++ {

	//go func(idx int) {
	_, err = mongo.InsertCollection(context.TODO(), fmt.Sprintf("hello_%d", 0))
	if err != nil {
		t.Fail()
	}
	//		wg.Done()
	//	}(i)
	//}
	//wg.Wait()

}
