package integration

import (
	"fmt"
	"github.com/knowhunger/ortoo/commons"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
	"github.com/knowhunger/ortoo/integration_test/test_helper"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"sync"
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

		fmt.Printf("Subscribed by client2\n")
		wg := sync.WaitGroup{}
		wg.Add(3)
		intCounter2 := client2.SubscribeIntCounter(key, commons.NewIntCounterHandlers(
			func(intCounter commons.IntCounter, old model.StateOfDatatype, new model.StateOfDatatype) {
				log.Logger.Infof("STATE: %s -> %s %d", old, new, intCounter.Get())
				require.Equal(n.T(), int32(1), intCounter.Get())
				wg.Done() // one time
			},
			func(intCounter commons.IntCounter, opList []interface{}) {
				for _, op := range opList {
					log.Logger.Infof("OPERATION: %+v", op)
				}
				wg.Done() // two times
			},
			func(err ...error) {
				require.NoError(n.T(), err[0])
			}))
		require.NoError(n.T(), client2.Sync())

		_, _ = intCounter1.IncreaseBy(10)
		require.NoError(n.T(), client1.Sync())
		wg.Wait()
		require.Equal(n.T(), intCounter1.Get(), intCounter2.Get())
	})
}
