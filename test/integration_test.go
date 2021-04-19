package integration

import (
	gocontext "context"
	context "github.com/knowhunger/ortoo/pkg/context"
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/server/mongodb"
	"github.com/knowhunger/ortoo/server/server"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"strings"
	"testing"
	"time"
)

const (
	dbName  = "test"
	TagTest = "TEST"
)

// IntegrationTestSuite is the base test suite for integration test.
type IntegrationTestSuite struct {
	suite.Suite
	collectionName string
	collectionNum  uint32
	server         *server.OrtooServer
	mongo          *mongodb.RepositoryMongo
	ctx            context.OrtooContext
}

// SetupTest builds some prerequisite for testing.
func (its *IntegrationTestSuite) SetupSuite() {
	its.ctx = context.NewOrtooContext(gocontext.TODO(), TagTest, context.MakeTagInTest(its.T().Name()))
	var err errors.OrtooError
	its.server, err = server.NewOrtooServer(gocontext.TODO(), NewTestOrtooServerConfig(dbName))
	if err != nil {
		its.Fail("fail to setup")
	}
	its.mongo = its.server.Mongo
	go func() {
		require.NoError(its.T(), its.server.Start())
	}()
	time.Sleep(1 * time.Second)
}

func (its *IntegrationTestSuite) SetupTest() {
	its.collectionName = strings.Split(its.T().Name(), "/")[1]
	its.ctx.L().Infof("set collection: %s", its.collectionName)
	var err error
	require.NoError(its.T(), its.mongo.PurgeCollection(its.ctx, its.collectionName))
	its.collectionNum, err = mongodb.MakeCollection(its.ctx, its.server.Mongo, its.collectionName)
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
