package integration

import (
	"github.com/knowhunger/ortoo/client"
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
		client1, err := client.NewOrtooClient(config)
		if err != nil {
			s.T().Fail()
			return
		}
		if err := client1.Connect(); err != nil {
			s.Suite.Fail("fail to connect server")
		}
		defer client1.Close()

		client1.Send()
	})

}

func (s *ClientServerTestSuite) TearDownTest() {
	s.server.Close()
	s.T().Log("TearDownTest")
}

func TestClientServerTestSuite(t *testing.T) {
	suite.Run(t, new(ClientServerTestSuite))
}
