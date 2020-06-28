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
		E1 string
		E2 int
		A3 []interface{}
	}{
		E1: "hello",
		E2: 1234,
		A3: arr,
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

		_, _ = docu1.PutToObject("E1", "X")
		_, _ = docu1.PutToObject("O2", strt1)
		_, _ = docu1.PutToObject("A3", arr1)
		docu2, err := docu1.GetFromObject("O2")
		require.NoError(its.T(), err)
		require.Equal(its.T(), docu2.GetDocumentType(), ortoo.TypeJSONObject)
		log.Logger.Infof("%v => %v", docu1.GetAsJSON(), docu2.GetAsJSON())

		docu2.PutToObject("E4", "world")

		docu3, err := docu1.GetFromObject("A3")
		require.Equal(its.T(), docu3.GetDocumentType(), ortoo.TypeJSONArray)
		docu3.InsertToArray(1, strt1)

		docu4, err := docu3.GetFromArray(1)
		require.NoError(its.T(), err)
		require.Equal(its.T(), docu4.GetDocumentType(), ortoo.TypeJSONObject)
		docu4.PutToObject("O4", strt1)
		log.Logger.Infof("Before DELETE:%v", docu1.GetAsJSON())
		docu1.DeleteInObject("E1")
		log.Logger.Infof("After DELETE E1:%v", docu1.GetAsJSON())
		docu1.DeleteInObject("O2")
		log.Logger.Infof("After DELETE O2:%v", docu1.GetAsJSON())
		docu3.DeleteInArray(1)
		log.Logger.Infof("After DELETE A3[1]:%v", docu1.GetAsJSON())

		require.NoError(its.T(), client1.Sync())

	})
}
