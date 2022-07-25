package integration

import (
	"github.com/orda-io/orda/client/pkg/errors"
	"github.com/orda-io/orda/client/pkg/model"
	orda2 "github.com/orda-io/orda/client/pkg/orda"
	"sync"

	"github.com/stretchr/testify/require"
)

func (its *IntegrationTestSuite) TestProtocol() {
	its.Run("Can produce an error when key is duplicated", func() {
		key := GetFunctionName()

		config := NewTestOrdaClientConfig(its.collectionName, model.SyncType_MANUALLY)

		client1 := orda2.NewClient(config, "client1")

		err := client1.Connect()
		require.NoError(its.T(), err)
		defer func() {
			_ = client1.Close()
		}()

		client2 := orda2.NewClient(config, "client2")
		err = client2.Connect()
		require.NoError(its.T(), err)
		defer func() {
			_ = client2.Close()
		}()

		_ = client1.CreateCounter(key, orda2.NewHandlers(
			func(dt orda2.Datatype, old model.StateOfDatatype, new model.StateOfDatatype) {
				require.Equal(its.T(), model.StateOfDatatype_DUE_TO_CREATE, old)
				require.Equal(its.T(), model.StateOfDatatype_SUBSCRIBED, new)
			}, nil,
			func(dt orda2.Datatype, errs ...errors.OrdaError) {
				require.NoError(its.T(), errs[0])
			}))
		require.NoError(its.T(), client1.Sync())
		wg := &sync.WaitGroup{}
		wg.Add(1)
		_ = client2.CreateCounter(key, orda2.NewHandlers(
			nil, nil,
			func(dt orda2.Datatype, errs ...errors.OrdaError) {
				its.ctx.L().Infof("should be duplicate error:%v", errs[0])
				require.Error(its.T(), errs[0])
				wg.Done()
			}))
		require.NoError(its.T(), client2.Sync())
		wg.Wait()
	})

	its.Run("Can produce RPC error when connect", func() {
		config := NewTestOrdaClientConfig("NOT_EXISTING", model.SyncType_MANUALLY)
		client1 := orda2.NewClient(config, its.getTestName())
		err := client1.Connect()
		require.Error(its.T(), err)
		defer func() {
			_ = client1.Close()
		}()
	})
}
