package integration

import (
	"github.com/knowhunger/ortoo/ortoo"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/ortoo/model"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

type ClientServerTestSuite struct {
	OrtooServerTestSuite
}

func (s *ClientServerTestSuite) SetupTest() {
	s.OrtooServerTestSuite.SetupTest()
}

func (s *ClientServerTestSuite) TestClientServer() {

	key := GetFunctionName()

	s.Run("Can create a client and a datatype with server", func() {
		config := NewTestOrtooClientConfig(s.collectionName)
		client1, err := ortoo.NewOrtooClient(config, "client1")
		require.NoError(s.T(), err)
		err = client1.Connect()
		require.NoError(s.T(), err)
		defer client1.Close()

		// intCounterCh1, errCh1 :=
		client1.CreateIntCounter(key, ortoo.NewIntCounterHandlers(
			func(intCounter ortoo.IntCounter, oldState, newState model.StateOfDatatype) {
				_, _ = intCounter.Increase()
				_, _ = intCounter.Increase()
				_, _ = intCounter.Increase()
				require.NoError(s.T(), client1.Sync())
			}, nil,
			func(errs ...error) {
				s.T().Fatal(errs[0])
			}))

		// intCounter1.DoTransaction("transaction1", func(counter ortoo.IntCounterInTxn) error {
		//	return nil
		// })

	})

	s.Run("Can subscribe the datatype", func() {
		config := NewTestOrtooClientConfig(s.collectionName)
		client2, err := ortoo.NewOrtooClient(config, "client2")
		if err != nil {
			s.T().Fatal(err)
		}
		if err := client2.Connect(); err != nil {
			s.T().Fatal(err)
		}
		defer client2.Close()

		client2.SubscribeIntCounter(key, ortoo.NewIntCounterHandlers(
			func(intCounter ortoo.IntCounter, oldState, newState model.StateOfDatatype) {
				log.Logger.Infof("%d", intCounter.Get())
				_, _ = intCounter.IncreaseBy(3)
				require.NoError(s.T(), client2.Sync())
			}, nil,
			func(errs ...error) {
				s.T().Fatal(errs[0])
			}))
	})
}

func TestClientServerTestSuite(t *testing.T) {
	suite.Run(t, new(ClientServerTestSuite))
}
