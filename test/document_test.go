package integration

import (
	"fmt"
	"github.com/orda-io/orda/client/pkg/errors"
	"github.com/orda-io/orda/client/pkg/log"
	"github.com/orda-io/orda/client/pkg/model"
	"github.com/orda-io/orda/client/pkg/orda"
	"github.com/orda-io/orda/client/pkg/testonly"
	"github.com/orda-io/orda/server/mongodb"
	"time"

	"github.com/stretchr/testify/require"
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

	handler = orda.NewHandlers(
		func(dt orda.Datatype, old model.StateOfDatatype, new model.StateOfDatatype) {

		},
		func(dt orda.Datatype, opList []interface{}) {

		},
		func(dt orda.Datatype, errs ...errors.OrdaError) {

		})
)

func (its *IntegrationTestSuite) TestDocument() {

	its.Run("Can exploit OPX", func() {
		config := NewTestOrdaClientConfig(its.collectionName, model.SyncType_MANUALLY)
		client1 := orda.NewClient(config, "docClient1")
		client2 := orda.NewClient(config, "docClient2")

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

		log.Logger.Infof("DOC1:%v", doc1.ToJSON())
		log.Logger.Infof("DOC2:%v", doc2.ToJSON())
		require.Equal(its.T(), doc1.ToJSON(), doc2.ToJSON())

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

	its.Run("Can store real snapshot for Document", func() {
		config := NewTestOrdaClientConfig(its.collectionName, model.SyncType_MANUALLY)
		client1 := orda.NewClient(config, its.getTestName())

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
		log.Logger.Infof("%v", testonly.Marshal(its.T(), doc1.ToJSON()))

		// Check the child Document
		require.Equal(its.T(), doc2.GetTypeOfJSON(), orda.TypeJSONObject)

		// Can put a new value to the child Document
		old4, err := doc2.PutToObject("E3", "world")
		require.NoError(its.T(), err)
		require.Nil(its.T(), old4)

		// Can obtain JSONArray: doc3 = ["world",1234,3.14]
		doc3, err := doc1.GetFromObject("A3")
		require.NoError(its.T(), err)
		require.Equal(its.T(), doc3.GetTypeOfJSON(), orda.TypeJSONArray)

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
			if snap != nil && snap[mongodb.Ver] == int64(6) {
				log.Logger.Infof("%v", testonly.Marshal(its.T(), snap))
				return true
			}
			return false
		}, 10*time.Second, 1*time.Second, "cannot make snapshot")
	})

	its.Run("Can test a transaction for Document", func() {
		config := NewTestOrdaClientConfig(its.collectionName, model.SyncType_MANUALLY)
		client1 := orda.NewClient(config, "docClient1")
		client2 := orda.NewClient(config, "docClient2")

		oErr := client1.Connect()
		require.NoError(its.T(), oErr)
		oErr = client2.Connect()
		require.NoError(its.T(), oErr)
		defer func() {
			_ = client1.Close()
			_ = client2.Close()
		}()

		doc1 := client1.CreateDocument(its.getTestName(), handler)
		_, _ = doc1.PutToObject("1", "a")
		require.NoError(its.T(), client1.Sync())
		doc2 := client2.SubscribeDocument(its.getTestName(), handler)
		require.NoError(its.T(), client2.Sync())

		err := doc1.Transaction("T1", func(document orda.DocumentInTx) error {
			_, _ = document.PutToObject("K1", "V1")
			_, _ = document.PutToObject("K2", str1)
			_, _ = document.PutToObject("K3", arr1)
			return nil
		})
		require.NoError(its.T(), err)
		require.NoError(its.T(), client1.Sync())
		require.NoError(its.T(), client2.Sync())
		its.ctx.L().Infof("doc1:%v", testonly.Marshal(its.T(), doc1.ToJSON()))
		its.ctx.L().Infof("doc2:%v", testonly.Marshal(its.T(), doc2.ToJSON()))

		err = doc2.Transaction("T2", func(document orda.DocumentInTx) error {
			_, _ = document.PutToObject("K1", "V2")
			k3, _ := document.GetFromObject("K3")
			_, _ = k3.InsertToArray(0, "V3")
			return fmt.Errorf("error")
		})

		its.ctx.L().Infof("doc1:%v", testonly.Marshal(its.T(), doc1.ToJSON()))
		its.ctx.L().Infof("doc2:%v", testonly.Marshal(its.T(), doc2.ToJSON()))
	})
}
