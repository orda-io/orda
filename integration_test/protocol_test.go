package integration

import (
	"github.com/knowhunger/ortoo/commons"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

type ProtocolTestSuite struct {
	OrtooServerTestSuite
}

func (n *ProtocolTestSuite) SetupTest() {
	n.OrtooServerTestSuite.SetupTest()
}

func TestProtocolTest(t *testing.T) {
	suite.Run(t, new(ProtocolTestSuite))
}

func (n *ProtocolTestSuite) TestProtocol() {
	key := GetFunctionName()

	n.Run("Can return duplicate key error for datatype", func() {
		config := NewTestOrtooClientConfig(n.collectionName)

		client1, err := commons.NewOrtooClient(config, "client1")
		require.NoError(n.T(), err)
		err = client1.Connect()
		require.NoError(n.T(), err)
		defer func() {
			_ = client1.Close()
		}()

		client2, err := commons.NewOrtooClient(config, "client2")
		require.NoError(n.T(), err)
		err = client2.Connect()
		require.NoError(n.T(), err)
		defer func() {
			_ = client2.Close()
		}()

		_ = client1.CreateIntCounter(key, commons.NewIntCounterHandlers(
			func(intCounter commons.IntCounter, old model.StateOfDatatype, new model.StateOfDatatype) {
				require.Equal(n.T(), model.StateOfDatatype_DUE_TO_CREATE, old)
				require.Equal(n.T(), model.StateOfDatatype_SUBSCRIBED, new)
			}, nil,
			func(errs ...error) {
				require.NoError(n.T(), errs[0])
			}))
		require.NoError(n.T(), client1.Sync())

		_ = client2.CreateIntCounter(key, commons.NewIntCounterHandlers(
			nil, nil,
			func(errs ...error) {
				log.Logger.Errorf("should be duplicate error:%v", errs[0])
				require.Error(n.T(), errs[0])
			}))
	})
}
