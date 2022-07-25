package integration

import (
	"github.com/orda-io/orda/client/pkg/errors"
	"github.com/orda-io/orda/client/pkg/model"
	orda2 "github.com/orda-io/orda/client/pkg/orda"
	"github.com/stretchr/testify/require"
	"time"
)

func (its *IntegrationTestSuite) TestList() {
	key := GetFunctionName()

	its.Run("Can update snapshot for list", func() {
		config := NewTestOrdaClientConfig(its.collectionName, model.SyncType_MANUALLY)
		client1 := orda2.NewClient(config, "listClient")

		err := client1.Connect()
		require.NoError(its.T(), err)
		its.ctx.L().Infof("end")
		defer func() {
			require.NoError(its.T(), client1.Close())
		}()
		// time.Sleep(3 * time.Second)
		list1 := client1.CreateList(key, orda2.NewHandlers(
			func(dt orda2.Datatype, old model.StateOfDatatype, new model.StateOfDatatype) {

			},
			func(dt orda2.Datatype, opList []interface{}) {

			},
			func(dt orda2.Datatype, errs ...errors.OrdaError) {

			}))
		_, _ = list1.InsertMany(0, "a", 2, 3.141592, time.Now())
		require.NoError(its.T(), client1.Sync())
		time.Sleep(2 * time.Second)
	})
}
