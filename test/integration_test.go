package integration

import (
	"context"
	"github.com/knowhunger/ortoo/pkg/log"
	"github.com/knowhunger/ortoo/server/mongodb"
	"github.com/knowhunger/ortoo/server/server"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"strings"
	"testing"
	"time"
)

const dbName = "test"

// IntegrationTestSuite is the base test suite for integration test.
type IntegrationTestSuite struct {
	suite.Suite
	collectionName string
	collectionNum  uint32
	server         *server.OrtooServer
	mongo          *mongodb.RepositoryMongo
}

// SetupTest builds some prerequisite for testing.
func (its *IntegrationTestSuite) SetupSuite() {

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

func (its *IntegrationTestSuite) SetupTest() {
	its.collectionName = strings.Split(its.T().Name(), "/")[1]
	log.Logger.Infof("Setup OrtooIntegrationTest:%s", its.collectionName)
	var err error
	require.NoError(its.T(), its.mongo.ResetCollections(context.Background(), its.collectionName))
	its.collectionNum, err = mongodb.MakeCollection(its.server.Mongo, its.collectionName)
	require.NoError(its.T(), err)
}

// TearDownTest tears down IntegrationTestSuite.
func (its *IntegrationTestSuite) TearDownSuite() {
	its.T().Log("TearDown IntegrationTestSuite")
	its.server.Close(true)
}

func (its *IntegrationTestSuite) getTestName() string {
	sp := strings.Split(its.T().Name(), "/")
	return sp[len(sp)-1]
}

func TestIntegration(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
