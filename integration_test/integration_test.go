package integration

import (
	"context"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/server"
	"github.com/knowhunger/ortoo/server/mongodb"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"strings"
	"testing"
	"time"
)

const dbName = "integration_test"

// OrtooIntegrationTestSuite is the base test suite for integration test.
type OrtooIntegrationTestSuite struct {
	suite.Suite
	collectionName string
	collectionNum  uint32
	server         *server.OrtooServer
	mongo          *mongodb.RepositoryMongo
}

// SetupTest builds some prerequisite for testing.
func (its *OrtooIntegrationTestSuite) SetupSuite() {

	var err error
	its.mongo, err = GetMongo(dbName)
	if err != nil {
		its.T().Fatal("fail to initialize mongoDB")
	}

	its.server, err = server.NewOrtooServer(context.TODO(), NewTestOrtooServerConfig(dbName))
	if err != nil {
		_ = log.OrtooError(err)
		its.Fail("fail to setup")
	}

	go func() {
		require.NoError(its.T(), its.server.Start())
	}()
	time.Sleep(1 * time.Second)
}

func (its *OrtooIntegrationTestSuite) SetupTest() {
	its.collectionName = strings.Split(its.T().Name(), "/")[1]
	log.Logger.Infof("Setup OrtooIntegrationTest:%s", its.collectionName)
	var err error
	require.NoError(its.T(), its.mongo.PurgeAllDocumentsOfCollection(context.Background(), its.collectionName))
	its.collectionNum, err = MakeTestCollection(its.server.Mongo, its.collectionName)
	require.NoError(its.T(), err)
}

// TearDownTest tears down OrtooIntegrationTestSuite.
func (its *OrtooIntegrationTestSuite) TearDownSuite() {
	its.T().Log("TearDown OrtooIntegrationTestSuite")
	its.server.Close()
}

func TestIntegration(t *testing.T) {
	suite.Run(t, new(OrtooIntegrationTestSuite))
}
