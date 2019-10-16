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
					return
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

	t.Run("Can get clientDoc with checkpoint", func(t *testing.T) {

	})

	t.Run("Can manipulate clientDoc", func(t *testing.T) {
		c := &schema.ClientDoc{
			CUID:          "test_cuid",
			Alias:         "test_alias",
			CollectionNum: 1,
			SyncType:      "MANUAL",
			CheckPoints: map[string]*model.CheckPoint{
				"test_duid1": model.NewCheckPoint().Set(1, 2),
				"test_duid2": model.NewCheckPoint().Set(3, 4),
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := mongo.UpdateClient(context.TODO(), c); err != nil {
			t.Fatal(err)
		}
		clientWithoutCheckPoints, err := mongo.GetClientWithoutCheckPoints(context.TODO(), c.CUID)
		if err != nil {
			t.Fatal(err)
		}

		clientWithCheckPoints, err := mongo.GetClient(context.TODO(), c.CUID)
		if err != nil {
			t.Fatal(err)
		}

		if err := mongo.DeleteClient(context.TODO(), c.CUID); err != nil {
			t.Fatal(err)
		}
		if err := mongo.DeleteClient(context.TODO(), c.CUID); err == nil {
			t.Fatal(err)
		}

		log.Logger.Infof("%+v", clientWithoutCheckPoints)
		log.Logger.Infof("%+v", clientWithCheckPoints)
	})

	t.Run("Can manipulate datatypeDoc", func(t *testing.T) {
		d := &schema.DatatypeDoc{
			DUID:          "test_duid",
			Key:           "test_key",
			CollectionNum: 1,
			Type:          "test_datatype",
			Sseq:          0,
			CreatedAt:     time.Now(),
		}
		if err := mongo.UpdateDatatype(context.TODO(), d); err != nil {
			t.Fatal(err)
		}
	})
}
