package integration

import (
	"github.com/knowhunger/ortoo/ortoo"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/ortoo/model"
	"github.com/stretchr/testify/require"
)

func (its *OrtooIntegrationTestSuite) TestDocument() {
	key := GetFunctionName()

	var strt1 = struct {
		K1 string
		K2 int
	}{
		K1: "hello",
		K2: 1234,
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

		_, _ = docu1.Add("K1", "X")
		_, _ = docu1.Add("K2", strt1)
		_, _ = docu1.Add("K3", arr1)
		docu2, err := docu1.GetByKey("K2")
		require.NoError(its.T(), err)
		log.Logger.Infof("%v", docu2)
		docu2.Add("K2_2", "hello")

		require.NoError(its.T(), client1.Sync())
	})
}
