package integration

import (
	"github.com/knowhunger/ortoo/commons"
	"github.com/knowhunger/ortoo/integration_test/test_helper"
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

	key := test_helper.GetFunctionName()

	s.Run("Can create a client and a datatype with server", func() {
		config := NewTestOrtooClientConfig(s.collectionName)
		client1, err := commons.NewOrtooClient(config, "client1")
		if err != nil {
			s.T().Fatal(err)
		}
		if err := client1.Connect(); err != nil {
			s.T().Fatal(err)
		}
		defer client1.Close()

		intCounterCh1, errCh1 := client1.CreateIntCounter(key)
		var intCounter1 commons.IntCounter
		select {
		case intCounter1 = <-intCounterCh1:
			_, _ = intCounter1.Increase()
			_, _ = intCounter1.Increase()
			_, _ = intCounter1.Increase()
			client1.Sync()
		case err1 := <-errCh1:
			s.T().Fatal(err1)
		}
	})

	// s.Run("Can subscribe the datatype", func() {
	// 	config := NewTestOrtooClientConfig(s.collectionName)
	// 	client2, err := commons.NewOrtooClient(config, "client2")
	// 	if err != nil {
	// 		s.T().Fatal(err)
	// 	}
	// 	if err := client2.Connect(); err != nil {
	// 		s.T().Fatal(err)
	// 	}
	// 	defer client2.Close()
	//
	// 	intCounterCh2, errCh2 := client2.SubscribeIntCounter(key)
	// 	var intCounter2 commons.IntCounter
	// 	select {
	// 	case intCounter2 = <-intCounterCh2:
	// 		log.Logger.Infof("%d", intCounter2.Get())
	// 	case err2 := <-errCh2:
	// 		s.T().Fatal(err2)
	// 	}
	// })
}

func TestClientServerTestSuite(t *testing.T) {
	suite.Run(t, new(ClientServerTestSuite))
}
