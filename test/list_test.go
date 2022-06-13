package integration

import (
	"github.com/orda-io/orda/pkg/errors"
	"github.com/orda-io/orda/pkg/model"
	"github.com/orda-io/orda/pkg/orda"
	"github.com/stretchr/testify/require"
	"time"
)

func (its *IntegrationTestSuite) TestList() {
	key := GetFunctionName()

	its.Run("Can update snapshot for list", func() {
		config := NewTestOrdaClientConfig(its.collectionName, model.SyncType_MANUALLY)
		client1 := orda.NewClient(config, "listClient")

		err := client1.Connect()
		require.NoError(its.T(), err)
		its.ctx.L().Infof("end")
		defer func() {
			require.NoError(its.T(), client1.Close())
		}()
		// time.Sleep(3 * time.Second)
		list1 := client1.CreateList(key, orda.NewHandlers(
			func(dt orda.Datatype, old model.StateOfDatatype, new model.StateOfDatatype) {

			},
			func(dt orda.Datatype, opList []interface{}) {

			},
			func(dt orda.Datatype, errs ...errors.OrdaError) {

			}))
		_, _ = list1.InsertMany(0, "a", 2, 3.141592, time.Now())
		require.NoError(its.T(), client1.Sync())
		time.Sleep(2 * time.Second)
	})
}
