package integration

import (
	"context"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/server"
	"github.com/knowhunger/ortoo/server/mongodb"
	"github.com/stretchr/testify/suite"
	"time"
)

const dbName = "integration_test"

// OrtooServerTestSuite is the base test suite for integration test.
type OrtooServerTestSuite struct {
	suite.Suite
	collectionName string
	collectionNum  uint32
	server         *server.OrtooServer
	mongo          *mongodb.RepositoryMongo
}

// SetupTest builds some prerequisite for testing.
func (o *OrtooServerTestSuite) SetupTest() {
	o.T().Logf("Setup OrtooServerTestSuite:%s", GetFileName())
	o.collectionName = GetFileName()
	var err error
	o.mongo, err = GetMongo(dbName)
	if err != nil {
		o.T().Fatal("fail to initialize mongoDB")
	}

	o.mongo.PurgeAllDocumentsOfCollection(context.Background(), o.collectionName)

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
	time.Sleep(1 * time.Second)
}

// TearDownTest tears down OrtooServerTestSuite.
func (o *OrtooServerTestSuite) TearDownTest() {
	o.T().Log("TearDown OrtooServerTestSuite")
	o.server.Close()
}
