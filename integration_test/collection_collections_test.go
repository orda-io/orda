package integration

import (
	"context"
	"fmt"
	"github.com/knowhunger/ortoo/server/mongodb"
	"github.com/knowhunger/ortoo/server/mongodb/schema"
	"log"
	"sync"
	"testing"
)

func TestMongo(t *testing.T) {
	t.Run("Make collection simultaneously", func(t *testing.T) {
		conf := NewTestMongoDBConfig("ortoo_unit_test")
		mongo, err := mongodb.New(context.TODO(), conf)
		if err != nil {
			log.Fatalf("fail to create mongoDB instance:%v", err)
		}

		madeCollections := make(map[uint32]*schema.CollectionDoc)
		wg := sync.WaitGroup{}
		wg.Add(10)
		for i := 0; i < 10; i++ {
			go func(idx int) {
				collection, err := mongo.InsertCollection(context.TODO(), fmt.Sprintf("hello_%d", idx))
				if err != nil {
					t.Fail()
				}
				madeCollections[collection.Num] = collection
				wg.Done()
			}(i)
		}

		wg.Wait()
		if len(madeCollections) != 10 {
			t.Fail()
		}
		for _, v := range madeCollections {
			if err := mongo.DeleteCollection(context.TODO(), v.Name); err != nil {
				t.Fail()
			}
		}
	})

}
