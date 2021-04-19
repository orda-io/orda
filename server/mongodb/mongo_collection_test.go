package mongodb_test

import (
	gocontext "context"
	"encoding/json"
	"fmt"
	"github.com/knowhunger/ortoo/pkg/iface"
	"github.com/knowhunger/ortoo/pkg/log"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/pkg/operations"
	"github.com/knowhunger/ortoo/pkg/types"
	"github.com/knowhunger/ortoo/server/constants"
	"github.com/knowhunger/ortoo/server/mongodb/schema"
	"github.com/knowhunger/ortoo/server/svrcontext"
	"github.com/knowhunger/ortoo/test"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"gotest.tools/assert"

	"sync"
	"testing"
	"time"
)

func TestMongo(t *testing.T) {
	ctx := svrcontext.NewServerContext(gocontext.TODO(), constants.TagTest).UpdateCollection(t.Name())
	mongo, err := integration.GetMongo(ctx, "ortoo_unit_test")
	if err != nil {
		t.Fatal("fail to initialize mongoDB")
	}

	t.Run("Can make collections simultaneously", func(t *testing.T) {
		madeCollections := make(map[uint32]*schema.CollectionDoc)

		for i := 0; i < 10; i++ {
			if err := mongo.DeleteCollection(ctx, fmt.Sprintf("hello_%d", i)); err != nil {
				t.Fail()
			}
		}

		wg := sync.WaitGroup{}
		wg.Add(10)
		for i := 0; i < 10; i++ {
			go func(idx int) {
				collection, err := mongo.InsertCollection(ctx, fmt.Sprintf("hello_%d", idx))
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

	})

	t.Run("Can manipulate clientDoc", func(t *testing.T) {
		c := &schema.ClientDoc{
			CUID:          "test_cuid",
			Alias:         "test_alias",
			CollectionNum: 1,
			Type:          0,
			SyncType:      0,
			CheckPoints: map[string]*model.CheckPoint{
				"test_duid1": model.NewCheckPoint().Set(1, 2),
				"test_duid2": model.NewCheckPoint().Set(3, 4),
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := mongo.UpdateClient(ctx, c); err != nil {
			t.Fatal(err)
		}
		clientWithoutCheckPoints, err := mongo.GetClientWithoutCheckPoints(ctx, c.CUID)
		if err != nil {
			t.Fatal(err)
		}

		err = mongo.UpdateCheckPointInClient(ctx, c.CUID, "test_duid1", &model.CheckPoint{Cseq: 2, Sseq: 2})
		if err != nil {
			t.Fatal(err)
		}

		clientWithCheckPoints, err := mongo.GetClient(ctx, c.CUID)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, clientWithCheckPoints.CheckPoints["test_duid1"].Sseq, uint64(2))

		if err := mongo.UnsubscribeDatatypeFromAllClients(ctx, "test_duid2"); err != nil {
			t.Fatal(err)
		}

		if err := mongo.UnsubscribeDatatypeFromClient(ctx, c.CUID, "test_duid1"); err != nil {
			t.Fatal(err)
		}
		clientWithCheckPoints2, err := mongo.GetClient(ctx, c.CUID)
		if err != nil {
			t.Fatal(err)
		}
		_, ok := clientWithCheckPoints2.CheckPoints["test_duid1"]
		assert.Equal(t, ok, false)

		if err := mongo.DeleteClient(ctx, c.CUID); err != nil {
			t.Fatal(err)
		}
		if err := mongo.DeleteClient(ctx, c.CUID); err != nil {
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
			Visible:       true,
			SseqEnd:       0,
			CreatedAt:     time.Now(),
		}
		if err := mongo.UpdateDatatype(ctx, d); err != nil {
			t.Fatal(err)
		}

		datatypeDoc1, err := mongo.GetDatatype(ctx, d.DUID)
		if err != nil {
			t.Fatal(err)
		}
		log.Logger.Infof("%+v", datatypeDoc1)
		datatypeDoc2, err := mongo.GetDatatype(ctx, "not exist")
		if err != nil {
			t.Fatal(err)
		}
		if datatypeDoc2 != nil {
			t.FailNow()
		}
		datatypeDoc3, err := mongo.GetDatatypeByKey(ctx, d.CollectionNum, d.Key)
		if err != nil {
			t.Fatal(err)
		}
		log.Logger.Infof("%+v", datatypeDoc3)
	})

	t.Run("Can manipulate operationDoc", func(t *testing.T) {
		snap, err := json.Marshal(&testSnapshot{Value: 1})
		require.NoError(t, err)
		op := operations.NewSnapshotOperation(
			model.TypeOfDatatype_COUNTER,
			model.StateOfDatatype_DUE_TO_CREATE,
			string(snap))

		op.ID = model.NewOperationIDWithCUID(types.NewUID())
		modelOp := op.ToModelOperation()
		// opb, _ := proto.Marshal(op)

		var operations []interface{}
		opDoc := schema.NewOperationDoc(modelOp, "test_duid", 1, 1)
		log.Logger.Infof("%+v", opDoc.GetOperation())
		log.Logger.Infof("%+v", modelOp)
		operations = append(operations, opDoc)

		_, err = mongo.DeleteOperation(ctx, opDoc.DUID, 1)
		if err != nil {
			t.Fatal(err)
		}

		if err := mongo.InsertOperations(ctx, operations); err != nil {
			t.Fatal(err)
		}

		opList, _, err := mongo.GetOperations(ctx, opDoc.DUID, 1, constants.InfinitySseq)
		if err != nil {
			t.Fatal(err)
		}
		require.Equal(t, 1, len(opList))

	})

	t.Run("Can change json to bson", func(t *testing.T) {
		j := &struct {
			Key   string
			Array []string
		}{
			Key:   "world",
			Array: []string{"x", "y"},
		}
		data1, err := bson.Marshal(j)
		require.NoError(t, err)
		log.Logger.Infof("%v", data1)
	})
}

type testSnapshot struct {
	Value int32 `json:"value"`
}

func (its *testSnapshot) SetBase(base iface.BaseDatatype) {
	panic("implement me")
}

func (its *testSnapshot) GetBase() iface.BaseDatatype {
	panic("implement me")
}

func (its *testSnapshot) GetAsJSONCompatible() interface{} {
	panic("implement me")
}

func (its *testSnapshot) GetAsJSON() interface{} {
	// j, err := json.Marshal(its)
	// if err != nil {
	// 	return "", errors.NewDatatypeError(errors.ErrDatatypeSnapshot, err.Error())
	// }
	return its
}

func (its *testSnapshot) CloneSnapshot() iface.Snapshot {
	return &testSnapshot{
		Value: its.Value,
	}
}
