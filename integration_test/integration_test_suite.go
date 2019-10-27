package integration

import (
	"context"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/integration_test/test_helper"
	"github.com/knowhunger/ortoo/server"
	"github.com/knowhunger/ortoo/server/mongodb"
	"github.com/stretchr/testify/suite"
)

const dbName = "integration_test"

type OrtooServerTestSuite struct {
	suite.Suite
	collectionName string
	collectionNum  uint32
	server         *server.OrtooServer
	mongo          *mongodb.RepositoryMongo
}

func (o *OrtooServerTestSuite) SetupTest() {
	o.T().Logf("Setup OrtooServerTestSuite:%s", test_helper.GetFileName())
	o.collectionName = test_helper.GetFileName()
	var err error
	o.mongo, err = test_helper.GetMongo(dbName)
	if err != nil {
		o.T().Fatal("fail to initialize mongoDB")
	}

	o.server, err = server.NewOrtooServer(context.TODO(), NewTestOrtooServerConfig(dbName))
	if err != nil {
		_ = log.OrtooError(err)
		o.Fail("fail to setup")
	}
	o.collectionNum, err = MakeTestCollection(o.server.Mongo, o.collectionName)
	if err != nil {
		o.Fail("fail to test collection")
	}
	go o.server.Start()
}

func (o *OrtooServerTestSuite) TearDownTest() {
	o.server.Close()
	o.T().Log("TearDown OrtooServerTestSuite")
}
