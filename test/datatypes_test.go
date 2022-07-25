package integration

import (
	errors2 "github.com/orda-io/orda/client/pkg/errors"
	"github.com/orda-io/orda/client/pkg/model"
	orda2 "github.com/orda-io/orda/client/pkg/orda"
	"sync"

	"github.com/stretchr/testify/require"
)

func (its *IntegrationTestSuite) TestClientServer() {
	key := GetFunctionName()

	its.Run("Can create a client and a datatype with server", func() {
		config := NewTestOrdaClientConfig(its.collectionName, model.SyncType_MANUALLY)
		client1 := orda2.NewClient(config, "client1")
		err := client1.Connect()
		require.NoError(its.T(), err)
		defer client1.Close()
		wg := sync.WaitGroup{}
		wg.Add(1)
		client1.CreateCounter(key, orda2.NewHandlers(
			func(dt orda2.Datatype, oldState, newState model.StateOfDatatype) {
				intCounter := dt.(orda2.Counter)
				_, _ = intCounter.Increase()
				_, _ = intCounter.Increase()
				_, _ = intCounter.Increase()
				require.NoError(its.T(), client1.Sync())
				wg.Done()
			}, nil,
			func(dt orda2.Datatype, errs ...errors2.OrdaError) {
				its.T().Fatal(errs[0])
			}))
		require.NoError(its.T(), client1.Sync())
		wg.Wait()
	})

	its.Run("Can subscribe not existing datatype", func() {
		config := NewTestOrdaClientConfig(its.collectionName, model.SyncType_MANUALLY)
		client2 := orda2.NewClient(config, "client2")
		err := client2.Connect()
		require.NoError(its.T(), err)
		defer func() {
			_ = client2.Close()
		}()
		wg := sync.WaitGroup{}
		wg.Add(1)
		client2.SubscribeCounter("NOT_EXISTING", orda2.NewHandlers(
			nil, nil,
			func(dt orda2.Datatype, errs ...errors2.OrdaError) {
				for _, ordaError := range errs {
					if ordaError.GetCode() == errors2.DatatypeSubscribe {
						wg.Done()
						return
					}
				}
				its.T().Fatal(errs[0])
			}))
		require.NoError(its.T(), client2.Sync())
		wg.Wait()
	})
}
