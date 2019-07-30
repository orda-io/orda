package integration_test

import (
	"github.com/knowhunger/ortoo/client"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/server"
	"github.com/stretchr/testify/suite"
	"testing"
)

type ClientServerTestSuite struct {
	suite.Suite
	server *server.OrtooServer
}

func (s *ClientServerTestSuite) SetupTest() {
	s.server = server.NewOrtooServer(server.DefaultConfig())
	go s.server.Start()
	s.T().Log("SetupTest")
}

func (s *ClientServerTestSuite) TestClientServer() {
	client := client.NewOrtooClient("127.0.0.1:19061")
	if err := client.Connect(); err != nil {
		_ = log.OrtooError(err, "fail to connect server")
	}
	defer client.Close()
	log.Logger.Infof("%+v", client)
}

func (s *ClientServerTestSuite) TearDownTest() {
	s.server.Close()
	s.T().Log("TearDownTest")
}

func TestClientServerTestSuite(t *testing.T) {
	suite.Run(t, new(ClientServerTestSuite))
}
