package integration

import (
	gocontext "context"
	"github.com/orda-io/orda/server/managers"
	"github.com/orda-io/orda/server/redis"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	context "github.com/orda-io/orda/pkg/context"
	"github.com/orda-io/orda/pkg/errors"
	"github.com/orda-io/orda/server/mongodb"
	"github.com/orda-io/orda/server/server"
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
	conf           *managers.OrdaServerConfig
	server         *server.OrdaServer
	mongo          *mongodb.RepositoryMongo
	redis          *redis.Client
	ctx            context.OrdaContext
}

// SetupTest builds some prerequisite for testing.
func (its *IntegrationTestSuite) SetupSuite() {
	its.ctx = context.NewOrdaContext(gocontext.TODO(), TagTest, context.MakeTagInTest(its.T().Name()))
	var err errors.OrdaError
	its.conf = NewTestOrdaServerConfig(dbName)

	if its.server, err = server.NewOrdaServer(gocontext.TODO(), its.conf); err != nil {
		its.Fail("fail to setup")
	}
	if err = its.setupClients(); err != nil {
		its.Fail("fail to setup client")
	}

	go func() {
		require.NoError(its.T(), its.server.Start())
	}()
	time.Sleep(1 * time.Second)
}

func (its *IntegrationTestSuite) setupClients() errors.OrdaError {
	var err errors.OrdaError
	if its.mongo, err = mongodb.New(its.ctx, &its.conf.Mongo); err != nil {
		return err
	}

	if its.redis, err = redis.New(its.ctx, &its.conf.Redis); err != nil {
		return err
	}
	return nil
}

func (its *IntegrationTestSuite) SetupTest() {
	its.collectionName = strings.Split(its.T().Name(), "/")[1]
	its.ctx.L().Infof("set collection: %s", its.collectionName)
	var err error
	require.NoError(its.T(), its.mongo.PurgeCollection(its.ctx, its.collectionName))
	its.collectionNum, err = mongodb.MakeCollection(its.ctx, its.mongo, its.collectionName)
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
