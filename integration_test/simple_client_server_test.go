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
		client1, err := ortoo.NewClient(config, "client1")
		require.NoError(its.T(), err)
		err = client1.Connect()
		require.NoError(its.T(), err)
		defer client1.Close()

		client1.CreateIntCounter(key, ortoo.NewIntCounterHandlers(
			func(intCounter ortoo.IntCounter, oldState, newState model.StateOfDatatype) {
				_, _ = intCounter.Increase()
				_, _ = intCounter.Increase()
				_, _ = intCounter.Increase()
				require.NoError(its.T(), client1.Sync())
			}, nil,
			func(errs ...error) {
				its.T().Fatal(errs[0])
			}))

	})

	its.Run("Can subscribe the datatype", func() {
		config := NewTestOrtooClientConfig(its.collectionName)
		client2, err := ortoo.NewClient(config, "client2")
		if err != nil {
			its.T().Fatal(err)
		}
		if err := client2.Connect(); err != nil {
			its.T().Fatal(err)
		}
		defer client2.Close()

		client2.SubscribeIntCounter(key, ortoo.NewIntCounterHandlers(
			func(intCounter ortoo.IntCounter, oldState, newState model.StateOfDatatype) {
				log.Logger.Infof("%d", intCounter.Get())
				_, _ = intCounter.IncreaseBy(3)
				require.NoError(its.T(), client2.Sync())
			}, nil,
			func(errs ...error) {
				its.T().Fatal(errs[0])
			}))
	})
}
