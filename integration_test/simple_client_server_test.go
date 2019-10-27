package integration

import (
	"context"
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

	s.mongo.PurgeDatatype(context.Background(), s.collectionNum, key)

	s.Run("Can create a client with server", func() {
		config := NewTestOrtooClientConfig(dbName, s.collectionName)
		client1, err := commons.NewOrtooClient(config)
		if err != nil {
			s.T().Fail()
			return
		}
		if err := client1.Connect(); err != nil {
			s.Suite.Fail("fail to connect server")
		}
		defer client1.Close()
		// intCounter1, err := commons.newIntCounter("key", client1)
		intCounterCh1, err1Ch := client1.CreateIntCounter(key)
		var intCounter1 commons.IntCounter
		select {
		case intCounter1 = <-intCounterCh1:
		case err1 := <-err1Ch:
			s.Suite.Fail("fail to :", err1)
		}
		if intCounter1 != nil {
			_, _ = intCounter1.Increase()
			_, _ = intCounter1.Increase()
			_, _ = intCounter1.Increase()
			client1.Sync()
		}
	})
}

func TestClientServerTestSuite(t *testing.T) {
	suite.Run(t, new(ClientServerTestSuite))
}
