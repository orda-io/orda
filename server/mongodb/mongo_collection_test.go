package mongodb_test

import (
	"encoding/json"
	"fmt"
	"github.com/orda-io/orda/client/pkg/log"
	"github.com/orda-io/orda/client/pkg/model"
	"github.com/orda-io/orda/client/pkg/operations"
	"github.com/orda-io/orda/client/pkg/types"
	"github.com/orda-io/orda/server/schema"
	"github.com/orda-io/orda/server/testonly"

	"github.com/orda-io/orda/server/constants"
	integration "github.com/orda-io/orda/test"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"

	"sync"
	"testing"
)

func TestMongo(t *testing.T) {
	mongo, ctx, collectionNum := integration.InitTestDBCollection(t, testonly.TestDBName)

	t.Run("Can make collections simultaneously", func(t *testing.T) {
		madeCollections := make(map[int32]*schema.CollectionDoc)

		for i := 0; i < 10; i++ {
			require.NoError(t, mongo.DeleteCollection(ctx, fmt.Sprintf("hello_%d", i)))
		}

		wg := sync.WaitGroup{}
		wg.Add(10)
		for i := 0; i < 10; i++ {
			go func(idx int) {
				collection, err := mongo.InsertCollection(ctx, fmt.Sprintf("hello_%d", idx))
				require.NoError(t, err)
				madeCollections[collection.Num] = collection
				ctx.L().Infof("made collection %d", collection.Num)
				wg.Done()
			}(i)
		}

		wg.Wait()
		require.Equal(t, 10, len(madeCollections))
	})

	t.Run("Can manipulate datatypeDoc", func(t *testing.T) {
		d := schema.NewDatatypeDoc("test_duid2", "test_key", collectionNum, "test_datatype")

		d.AddNewClient("aaaa", int8(model.ClientType_EPHEMERAL), true)

		require.NoError(t, mongo.UpdateDatatype(ctx, d))

		datatypeDoc1, err := mongo.GetDatatype(ctx, d.DUID)
		require.NoError(t, err)

		log.Logger.Infof("%v", datatypeDoc1)
		datatypeDoc2, err := mongo.GetDatatype(ctx, "not exist")
		require.NoError(t, err)
		require.Nil(t, datatypeDoc2)

		datatypeDoc3, err := mongo.GetDatatypeByKey(ctx, d.CollectionNum, d.Key)
		require.NoError(t, err)

		log.Logger.Infof("%+v", datatypeDoc3)
	})

	t.Run("Can manipulate operationDoc", func(t *testing.T) {
		snap, err := json.Marshal(&testSnapshot{Value: 1})
		require.NoError(t, err)
		op := operations.NewSnapshotOperation(model.TypeOfDatatype_DOCUMENT, snap)

		op.ID = model.NewOperationIDWithCUID(types.NewUID())
		modelOp := op.ToModelOperation()

		var oplist []interface{}
		opDoc := schema.NewOperationDoc(modelOp, "test_duid", 1, collectionNum)
		log.Logger.Infof("%+v", opDoc.GetOperation())
		log.Logger.Infof("%+v", modelOp)
		oplist = append(oplist, opDoc)

		_, err = mongo.DeleteOperation(ctx, opDoc.DUID, 1)
		if err != nil {
			t.Fatal(err)
		}

		if err := mongo.InsertOperations(ctx, oplist); err != nil {
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

func (its *testSnapshot) ToJSON() interface{} {
	return its
}
