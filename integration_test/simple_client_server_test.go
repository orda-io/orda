package integration

import (
	"context"
	"github.com/knowhunger/ortoo/commons"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/server"
	"github.com/stretchr/testify/suite"
	"testing"
)

const (
	dbName         = "ortoo_test_db"
	collectionName = "ortoo_test_collection"
)

type ClientServerTestSuite struct {
	suite.Suite
	server *server.OrtooServer
}

func (s *ClientServerTestSuite) SetupTest() {
	s.T().Log("SetupTest")
	var err error
	s.server, err = server.NewOrtooServer(context.TODO(), NewTestOrtooServerConfig(dbName))
	if err != nil {
		_ = log.OrtooError(err)
		s.Fail("fail to setup")
	}
	if err := MakeTestCollection(s.server.Mongo, collectionName); err != nil {
		s.Fail("fail to test collection")
	}
	go s.server.Start()

}

func (s *ClientServerTestSuite) TestClientServer() {
	s.Run("Can create a client with server", func() {
		config := NewTestOrtooClientConfig(dbName, collectionName)
		client1, err := commons.NewOrtooClient(config)
		if err != nil {
			s.T().Fail()
			return
		}
		if err := client1.Connect(); err != nil {
			s.Suite.Fail("fail to connect server")
		}
		defer client1.Close()
		//intCounter1, err := commons.newIntCounter("key", client1)
		intCounterCh1, err1Ch := client1.SubscribeOrCreateIntCounter("key")
		var intCounter1 commons.IntCounter
		select {
		case intCounter1 = <-intCounterCh1:
		case err1 := <-err1Ch:
			s.Suite.Fail("fail to :", err1)
		}
		intCounter1.Increase()
		client1.Sync()
	})

}

func (s *ClientServerTestSuite) TearDownTest() {
	s.server.Close()
	s.T().Log("TearDownTest")
}

func TestClientServerTestSuite(t *testing.T) {
	suite.Run(t, new(ClientServerTestSuite))
}
