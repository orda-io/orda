package integration

import (
	"github.com/knowhunger/ortoo/ortoo"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/ortoo/model"
	"github.com/stretchr/testify/require"
)

func (its *OrtooIntegrationTestSuite) TestDocument() {
	key := GetFunctionName()

	var arr = []interface{}{"a", 2}
	var strt1 = struct {
		K1 string
		K2 int
		K3 []interface{}
	}{
		K1: "hello",
		K2: 1234,
		K3: arr,
	}

	var arr1 = []interface{}{"world", 1234, 3.14}

	its.Run("Can update snapshot for list", func() {
		config := NewTestOrtooClientConfig(its.collectionName)
		client1 := ortoo.NewClient(config, "listClient")

		err := client1.Connect()
		require.NoError(its.T(), err)
		defer func() {
			_ = client1.Close()
		}()

		docu1 := client1.CreateDocument(key, ortoo.NewHandlers(
			func(dt ortoo.Datatype, old model.StateOfDatatype, new model.StateOfDatatype) {

			},
			func(dt ortoo.Datatype, opList []interface{}) {

			},
			func(dt ortoo.Datatype, errs ...error) {

			}))

		_, _ = docu1.Put("K1", "X")
		_, _ = docu1.Put("K2", strt1)
		_, _ = docu1.Put("K3", arr1)
		docu2, err := docu1.GetByKey("K2")
		require.NoError(its.T(), err)
		require.Equal(its.T(), docu2.GetDocumentType(), ortoo.TypeJSONObject)
		log.Logger.Infof("%v", docu2)
		docu2.Put("K4", "world")

		docu3, err := docu1.GetByKey("K3")
		require.Equal(its.T(), docu3.GetDocumentType(), ortoo.TypeJSONArray)

		docu4, err := docu3.GetByIndex(0)
		require.NoError(its.T(), err)
		require.Equal(its.T(), docu4.GetDocumentType(), ortoo.TypeJSONElement)

		require.NoError(its.T(), client1.Sync())
	})
}
