package integration

import (
	gocontext "context"
	context "github.com/orda-io/orda/pkg/context"
	"github.com/orda-io/orda/pkg/errors"
	"github.com/orda-io/orda/server/mongodb"
	"github.com/orda-io/orda/server/server"
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
	server         *server.OrdaServer
	mongo          *mongodb.RepositoryMongo
	ctx            context.OrdaContext
}

// SetupTest builds some prerequisite for testing.
func (its *IntegrationTestSuite) SetupSuite() {
	its.ctx = context.NewOrdaContext(gocontext.TODO(), TagTest, context.MakeTagInTest(its.T().Name()))
	var err errors.OrdaError
	its.server, err = server.NewOrdaServer(gocontext.TODO(), NewTestOrdaServerConfig(dbName))
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
