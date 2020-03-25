package integration

import (
	"github.com/knowhunger/ortoo/ortoo"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/ortoo/model"
	"github.com/stretchr/testify/require"
)

func (its *OrtooIntegrationTestSuite) TestClientServer() {
	key := GetFunctionName()

	its.Run("Can create a client and a datatype with server", func() {
		config := NewTestOrtooClientConfig(its.collectionName)
		client1 := ortoo.NewClient(config, "client1")
		err := client1.Connect()
		require.NoError(its.T(), err)
		defer client1.Close()

		client1.CreateIntCounter(key, ortoo.NewHandlers(
			func(dt ortoo.Datatype, oldState, newState model.StateOfDatatype) {
				intCounter := dt.(ortoo.IntCounter)
				_, _ = intCounter.Increase()
				_, _ = intCounter.Increase()
				_, _ = intCounter.Increase()
				require.NoError(its.T(), client1.Sync())
			}, nil,
			func(dt ortoo.Datatype, errs ...error) {
				its.T().Fatal(errs[0])
			}))
		require.NoError(its.T(), client1.Sync())

	})

	its.Run("Can subscribe the datatype", func() {
		config := NewTestOrtooClientConfig(its.collectionName)
		client2 := ortoo.NewClient(config, "client2")
		err := client2.Connect()
		require.NoError(its.T(), err)
		defer client2.Close()

		client2.SubscribeIntCounter(key, ortoo.NewHandlers(
			func(dt ortoo.Datatype, oldState, newState model.StateOfDatatype) {
				intCounter := dt.(ortoo.IntCounter)
				log.Logger.Infof("%d", intCounter.Get())
				_, _ = intCounter.IncreaseBy(3)
				require.NoError(its.T(), client2.Sync())
			}, nil,
			func(dt ortoo.Datatype, errs ...error) {
				its.T().Fatal(errs[0])
			}))
	})
}
