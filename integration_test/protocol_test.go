package integration

import (
	"github.com/knowhunger/ortoo/commons"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/integration_test/test_helper"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"sync"
	"testing"
)

type ProtocolTestSuite struct {
	OrtooServerTestSuite
}

func (n *ProtocolTestSuite) SetupTest() {
	n.OrtooServerTestSuite.SetupTest()
}

func TestProtocolTest(t *testing.T) {
	suite.Run(t, new(ProtocolTestSuite))
}

func (n *ProtocolTestSuite) TestProtocol() {
	key := test_helper.GetFunctionName()

	n.Run("Can return duplicate error", func() {
		config := NewTestOrtooClientConfig(n.collectionName)
		client1, err := commons.NewOrtooClient(config, "client1")
		require.NoError(n.T(), err)

		err = client1.Connect()
		require.NoError(n.T(), err)
		defer func() {
			_ = client1.Close()
		}()
		var wg1 = sync.WaitGroup{}
		wg1.Add(1)
		intCounter1 := client1.CreateIntCounter(key, commons.NewIntCounterHandlers(
			func(intCounter commons.IntCounter) {
				wg1.Done()
			},
			nil,
			func(err error) {
				require.NoError(n.T(), err)
			}))
		wg1.Wait()
		_, _ = intCounter1.IncreaseBy(1)
		_, _ = intCounter1.IncreaseBy(2)
		_, _ = intCounter1.IncreaseBy(3)
		require.NoError(n.T(), client1.Sync())

		client2, err := commons.NewOrtooClient(config, "client2")
		require.NoError(n.T(), err)
		err = client2.Connect()
		require.NoError(n.T(), err)
		defer func() {
			_ = client2.Close()
		}()

		wg2 := sync.WaitGroup{}
		wg2.Add(1)
		_ = client2.CreateIntCounter(key, commons.NewIntCounterHandlers(
			func(intCounter commons.IntCounter) {
			}, nil,
			func(err error) {
				log.Logger.Errorf("should be duplicate error:%v", err)
				require.Error(n.T(), err)
				wg2.Done()
			}))
		wg2.Wait()
	})
}
