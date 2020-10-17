package integration

import (
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/log"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/pkg/ortoo"
	"github.com/knowhunger/ortoo/pkg/testonly"
	"github.com/stretchr/testify/require"
	"time"
)

var (
	arr  = []interface{}{"a", 2}
	arr1 = []interface{}{"world", 1234, 3.14}

	str1 = struct {
		E1 string
		E2 int
		A3 []interface{}
	}{
		E1: "hello",
		E2: 1234,
		A3: arr,
	}

	json1 = struct {
		K1_1 struct {
			K1_1_1 string
		}
	}{
		K1_1: struct {
			K1_1_1 string
		}{
			K1_1_1: "E1_1_1",
		},
	}

	handler = ortoo.NewHandlers(
		func(dt ortoo.Datatype, old model.StateOfDatatype, new model.StateOfDatatype) {

		},
		func(dt ortoo.Datatype, opList []interface{}) {

		},
		func(dt ortoo.Datatype, errs ...errors.OrtooError) {

		})
)

func (its *IntegrationTestSuite) TestDocumentExploitCommutativity() {

	its.Run("Can exploit OPX", func() {
		config := NewTestOrtooClientConfig(its.collectionName)
		client1 := ortoo.NewClient(config, "docClient1")
		client2 := ortoo.NewClient(config, "docClient2")

		err := client1.Connect()
		require.NoError(its.T(), err)
		err = client2.Connect()
		require.NoError(its.T(), err)
		defer func() {
			_ = client1.Close()
			_ = client2.Close()
		}()

		doc1 := client1.SubscribeOrCreateDocument(its.getTestName(), handler)
		_, _ = doc1.PutToObject("K1", json1)
		require.NoError(its.T(), client1.Sync())
		doc2 := client2.SubscribeOrCreateDocument(its.getTestName(), handler)
		require.NoError(its.T(), client2.Sync())

		log.Logger.Infof("DOC1:%v", doc1.GetAsJSON())
		log.Logger.Infof("DOC2:%v", doc2.GetAsJSON())
		require.Equal(its.T(), doc1.GetAsJSON(), doc2.GetAsJSON())

		_, _ = doc1.PutToObject("K1", "hello")
		_, _ = doc2.PutToObject("K1", "world")
		log.Logger.Infof("sync1")
		require.NoError(its.T(), client1.Sync())

		// log.Logger.Infof("sync2")
		// require.NoError(its.T(), client2.Sync())
		// log.Logger.Infof("sync3")
		// require.NoError(its.T(), client1.Sync())
		// log.Logger.Infof("DOC1:%v", doc1.GetAsJSON())
		// log.Logger.Infof("DOC2:%v", doc2.GetAsJSON())
		// require.Equal(its.T(), doc1.GetAsJSON(), doc2.GetAsJSON())
	})
}

func (its *IntegrationTestSuite) TestDocumentRealSnapshot() {
	its.Run("Can store real snapshot for Document", func() {
		config := NewTestOrtooClientConfig(its.collectionName)
		client1 := ortoo.NewClient(config, its.getTestName())

		err := client1.Connect()
		require.NoError(its.T(), err)
		defer func() {
			_ = client1.Close()
		}()

		doc1 := client1.CreateDocument(its.getTestName(), handler)

		// {"E1":"X"}
		old1, err := doc1.PutToObject("E1", "X")
		require.NoError(its.T(), err)
		require.Nil(its.T(), old1)
		// {"E1":"hello","O2":{"A3":["a",2],"E2":1234}}
		old2, err := doc1.PutToObject("O2", str1)
		require.Nil(its.T(), old2)
		require.NoError(its.T(), err)
		// doc1: {"E1":"X","O2":{"A3":["a",2],"E1":"hello","E2":1234},"A3":["world",1234,3.14]}
		old3, err := doc1.PutToObject("A3", arr1)
		require.NoError(its.T(), err)
		require.Nil(its.T(), old3)

		// Can obtain JSONObject: doc2 = {"A3":["a",2],"E1":"hello","E2":1234}
		doc2, err := doc1.GetFromObject("O2")
		require.NoError(its.T(), err)
		log.Logger.Infof("%v", testonly.Marshal(its.T(), doc1.GetAsJSON()))

		// Check the child Document
		require.Equal(its.T(), doc2.GetJSONType(), ortoo.TypeJSONObject)

		// Can put a new value to the child Document
		old4, err := doc2.PutToObject("E3", "world")
		require.NoError(its.T(), err)
		require.Nil(its.T(), old4)

		// Can obtain JSONArray: doc3 = ["world",1234,3.14]
		doc3, err := doc1.GetFromObject("A3")
		require.NoError(its.T(), err)
		require.Equal(its.T(), doc3.GetJSONType(), ortoo.TypeJSONArray)

		// Can insert a new JSONObject to the child JSONArray Document
		doc3a, err := doc3.InsertToArray(1, str1)
		require.NoError(its.T(), err)
		// Should return the same JSONArray Document
		require.Equal(its.T(), doc3a, doc3)

		time.Sleep(3000)
		require.NoError(its.T(), client1.Sync())

		require.Eventually(its.T(), func() bool {
			snap, err := its.mongo.GetRealSnapshot(its.ctx, its.collectionName, its.getTestName())
			require.NoError(its.T(), err)
			if snap != nil && snap["_ver"] == int64(6) {
				log.Logger.Infof("%v", testonly.Marshal(its.T(), snap))
				return true
			}
			return false
		}, 10*time.Second, 1*time.Second, "cannot make snapshot")
	})

	its.Run("Can subscribe with snapshot", func() {

	})

}
