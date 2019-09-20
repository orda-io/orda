package integration

import (
	"github.com/knowhunger/ortoo/commons"
	"github.com/knowhunger/ortoo/server"
	"github.com/stretchr/testify/suite"
	"testing"
)

type ClientServerTestSuite struct {
	suite.Suite
	server *server.OrtooServer
}

func (s *ClientServerTestSuite) SetupTest() {
	s.server = server.NewOrtooServer(NewTestOrtooServerConfig())
	go s.server.Start()
	s.T().Log("SetupTest")
}

func (s *ClientServerTestSuite) TestClientServer() {
	s.Run("Can create a client with server", func() {
		config := NewTestOrtooClientConfig()
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
		intCounterCh1, err1Ch := client1.LinkIntCounter("key")
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
