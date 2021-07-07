package integration

import (
	"github.com/orda-io/orda/pkg/errors"
	"github.com/orda-io/orda/pkg/model"
	"github.com/orda-io/orda/pkg/orda"
	"github.com/stretchr/testify/require"
	"sync"
)

func (its *IntegrationTestSuite) TestProtocol() {
	its.Run("Can produce an error when key is duplicated", func() {
		key := GetFunctionName()

		config := NewTestOrdaClientConfig(its.collectionName, model.SyncType_MANUALLY)

		client1 := orda.NewClient(config, "client1")

		err := client1.Connect()
		require.NoError(its.T(), err)
		defer func() {
			_ = client1.Close()
		}()

		client2 := orda.NewClient(config, "client2")
		err = client2.Connect()
		require.NoError(its.T(), err)
		defer func() {
			_ = client2.Close()
		}()

		_ = client1.CreateCounter(key, orda.NewHandlers(
			func(dt orda.Datatype, old model.StateOfDatatype, new model.StateOfDatatype) {
				require.Equal(its.T(), model.StateOfDatatype_DUE_TO_CREATE, old)
				require.Equal(its.T(), model.StateOfDatatype_SUBSCRIBED, new)
			}, nil,
			func(dt orda.Datatype, errs ...errors.OrdaError) {
				require.NoError(its.T(), errs[0])
			}))
		require.NoError(its.T(), client1.Sync())
		wg := &sync.WaitGroup{}
		wg.Add(1)
		_ = client2.CreateCounter(key, orda.NewHandlers(
			nil, nil,
			func(dt orda.Datatype, errs ...errors.OrdaError) {
				its.ctx.L().Infof("should be duplicate error:%v", errs[0])
				require.Error(its.T(), errs[0])
				wg.Done()
			}))
		require.NoError(its.T(), client2.Sync())
		wg.Wait()
	})

	its.Run("Can produce RPC error when connect", func() {
		config := NewTestOrdaClientConfig("NOT_EXISTING", model.SyncType_MANUALLY)
		client1 := orda.NewClient(config, its.getTestName())
		err := client1.Connect()
		require.Error(its.T(), err)
		defer func() {
			_ = client1.Close()
		}()
	})
}
