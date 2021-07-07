package integration

import (
	"github.com/orda-io/orda/pkg/errors"
	"github.com/orda-io/orda/pkg/iface"
	"github.com/orda-io/orda/pkg/log"
	"github.com/orda-io/orda/pkg/model"
	"github.com/orda-io/orda/pkg/operations"
	"github.com/orda-io/orda/pkg/orda"
	"github.com/stretchr/testify/require"
)

func (its *IntegrationTestSuite) TestMap() {
	key := GetFunctionName()

	its.Run("Can update snapshot for hash map", func() {
		config := NewTestOrdaClientConfig(its.collectionName, model.SyncType_MANUALLY)
		client1 := orda.NewClient(config, "client1")

		err := client1.Connect()
		require.NoError(its.T(), err)
		defer func() {
			_ = client1.Close()
		}()

		map1 := client1.CreateMap(key, orda.NewHandlers(
			func(dt orda.Datatype, old model.StateOfDatatype, new model.StateOfDatatype) {

			},
			func(dt orda.Datatype, opList []interface{}) {

			},
			func(dt orda.Datatype, errs ...errors.OrdaError) {

			}))
		_, _ = map1.Put("hello", "world")
		_, _ = map1.Put("num", 1234)
		_, _ = map1.Put("float", 3.141592)
		_, _ = map1.Put("struct", struct {
			ID  string
			Age uint
		}{
			ID:  "hello",
			Age: 10,
		})
		_, _ = map1.Put("list", []string{"x", "y", "z"})
		_, _ = map1.Put("Removed", "deleted")
		_, _ = map1.Remove("Removed")
		require.Nil(its.T(), map1.Get("Removed"))
		require.NoError(its.T(), client1.Sync())
		sop, err := operations.NewSnapshotOperationFromDatatype(map1.(iface.Datatype))
		log.Logger.Infof("%v", sop.String())
	})
}
