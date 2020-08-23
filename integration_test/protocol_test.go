package integration

import (
	"github.com/knowhunger/ortoo/ortoo"
	"github.com/knowhunger/ortoo/ortoo/errors"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/ortoo/model"
	"github.com/stretchr/testify/require"
)

func (its *OrtooIntegrationTestSuite) TestProtocol() {
	key := GetFunctionName()

	its.Run("Can return duplicate key error for datatype", func() {
		config := NewTestOrtooClientConfig(its.collectionName)

		client1 := ortoo.NewClient(config, "client1")

		err := client1.Connect()
		require.NoError(its.T(), err)
		defer func() {
			_ = client1.Close()
		}()

		client2 := ortoo.NewClient(config, "client2")
		err = client2.Connect()
		require.NoError(its.T(), err)
		defer func() {
			_ = client2.Close()
		}()

		_ = client1.CreateCounter(key, ortoo.NewHandlers(
			func(dt ortoo.Datatype, old model.StateOfDatatype, new model.StateOfDatatype) {
				require.Equal(its.T(), model.StateOfDatatype_DUE_TO_CREATE, old)
				require.Equal(its.T(), model.StateOfDatatype_SUBSCRIBED, new)
			}, nil,
			func(dt ortoo.Datatype, errs ...errors.OrtooError) {
				require.NoError(its.T(), errs[0])
			}))
		require.NoError(its.T(), client1.Sync())

		_ = client2.CreateCounter(key, ortoo.NewHandlers(
			nil, nil,
			func(dt ortoo.Datatype, errs ...errors.OrtooError) {
				log.Logger.Errorf("should be duplicate error:%v", errs[0])
				require.Error(its.T(), errs[0])
			}))
	})
}
