package mongodb_test

import (
	"context"
	"fmt"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
	"github.com/knowhunger/ortoo/server/mongodb"
	"github.com/knowhunger/ortoo/server/mongodb/schema"
	//"log"
	"sync"
	"testing"
	"time"
)

func TestMongo(t *testing.T) {

	conf := mongodb.NewTestMongoDBConfig("ortoo_unit_test")
	mongo, err := mongodb.New(context.TODO(), conf)
	if err != nil {
		log.Logger.Fatalf("fail to create mongoDB instance:%v", err)
	}

	t.Run("Make collections simultaneously", func(t *testing.T) {
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

	t.Run("Can update clientDoc", func(t *testing.T) {
		c := &schema.ClientDoc{
			CUID:       "test_cuid",
			Alias:      "test_alias",
			Collection: "test_collection",
			SyncType:   "MANUAL",
			CheckPoints: map[string]*model.CheckPoint{
				"test_duid1": model.NewCheckPoint().Set(0, 1),
				"test_duid2": model.NewCheckPoint().Set(1, 0),
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := mongo.UpdateClient(context.TODO(), c); err != nil {
			t.Fail()
		}
		cstored, err := mongo.GetClient(context.TODO(), c.CUID)
		if err != nil {
			t.Fail()
		}
		log.Logger.Infof("%+v", cstored)
	})

}
