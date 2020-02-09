package integration

import (
	"github.com/knowhunger/ortoo/commons"
	"github.com/knowhunger/ortoo/integration_test/test_helper"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

type NotificationTestSuite struct {
	OrtooServerTestSuite
}

func (n *NotificationTestSuite) SetupTest() {
	n.OrtooServerTestSuite.SetupTest()
}

func Test1(t *testing.T) {
	suite.Run(t, new(NotificationTestSuite))
}

func (n *NotificationTestSuite) TestNotificationTest() {
	key := test_helper.GetFunctionName()
	n.Run("Can notify remote change", func() {
		config := NewTestOrtooClientConfig(n.collectionName)
		client1, err := commons.NewOrtooClient(config, "client1")
		require.NoError(n.T(), err)

		err = client1.Connect()
		require.NoError(n.T(), err)
		defer func() {
			_ = client1.Close()
		}()

		client2, err := commons.NewOrtooClient(config, "client2")
		require.NoError(n.T(), err)
		err = client2.Connect()
		require.NoError(n.T(), err)
		defer func() {
			_ = client2.Close()
		}()

		intCounter1 := client1.CreateIntCounter(key, nil)
		_, _ = intCounter1.Increase()
		require.NoError(n.T(), client1.Sync())

		// _ = client2.SubscribeIntCounter(key, commons.NewIntCounterHandlers(
		// 	func(intCounter commons.IntCounter) {
		// 		require.Equal(n.T(), int32(0), intCounter.Get())
		// 	},
		// 	func(intCounter commons.IntCounter, opList []model.Operation) {
		//
		// 	},
		// 	func(err error) {
		// 		require.NoError(n.T(), err)
		// 	}))
		// require.NoError(n.T(), client2.Sync())

		// require.NoError(n.T(), client1.Sync())
	})
}
