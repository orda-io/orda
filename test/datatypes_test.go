package integration

import (
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/log"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/pkg/ortoo"
	"github.com/stretchr/testify/require"
	"sync"
)

func (its *IntegrationTestSuite) TestClientServer() {
	key := GetFunctionName()

	its.Run("Can create a client and a datatype with server", func() {
		config := NewTestOrtooClientConfig(its.collectionName)
		client1 := ortoo.NewClient(config, "client1")
		err := client1.Connect()
		require.NoError(its.T(), err)
		defer client1.Close()
		wg := sync.WaitGroup{}
		wg.Add(1)
		client1.CreateCounter(key, ortoo.NewHandlers(
			func(dt ortoo.Datatype, oldState, newState model.StateOfDatatype) {
				intCounter := dt.(ortoo.Counter)
				_, _ = intCounter.Increase()
				_, _ = intCounter.Increase()
				_, _ = intCounter.Increase()
				require.NoError(its.T(), client1.Sync())
				wg.Done()
			}, nil,
			func(dt ortoo.Datatype, errs ...errors.OrtooError) {
				its.T().Fatal(errs[0])
			}))
		require.NoError(its.T(), client1.Sync())
		wg.Wait()
	})

	its.Run("Can subscribe the datatype", func() {
		config := NewTestOrtooClientConfig(its.collectionName)
		client2 := ortoo.NewClient(config, "client2")
		err := client2.Connect()
		require.NoError(its.T(), err)
		defer client2.Close()
		wg := sync.WaitGroup{}
		wg.Add(1)
		client2.SubscribeCounter(key, ortoo.NewHandlers(
			func(dt ortoo.Datatype, oldState, newState model.StateOfDatatype) {
				intCounter := dt.(ortoo.Counter)
				log.Logger.Infof("%d", intCounter.Get())
				_, _ = intCounter.IncreaseBy(3)
				require.NoError(its.T(), client2.Sync())
				wg.Done()
			}, nil,
			func(dt ortoo.Datatype, errs ...errors.OrtooError) {
				its.T().Fatal(errs[0])
			}))
		require.NoError(its.T(), client2.Sync())
		wg.Wait()
	})
}
