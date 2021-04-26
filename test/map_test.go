package integration

import (
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/pkg/ortoo"
	"github.com/stretchr/testify/require"
)

func (its *IntegrationTestSuite) TestMap() {
	key := GetFunctionName()

	its.Run("Can update snapshot for hash map", func() {
		config := NewTestOrtooClientConfig(its.collectionName, model.SyncType_MANUALLY)
		client1 := ortoo.NewClient(config, "client1")

		err := client1.Connect()
		require.NoError(its.T(), err)
		defer func() {
			_ = client1.Close()
		}()

		hashMap1 := client1.CreateMap(key, ortoo.NewHandlers(
			func(dt ortoo.Datatype, old model.StateOfDatatype, new model.StateOfDatatype) {

			},
			func(dt ortoo.Datatype, opList []interface{}) {

			},
			func(dt ortoo.Datatype, errs ...errors.OrtooError) {

			}))
		_, _ = hashMap1.Put("hello", "world")
		_, _ = hashMap1.Put("num", 1234)
		_, _ = hashMap1.Put("float", 3.141592)
		_, _ = hashMap1.Put("struct", struct {
			ID  string
			Age uint
		}{
			ID:  "hello",
			Age: 10,
		})
		_, _ = hashMap1.Put("list", []string{"x", "y", "z"})
		_, _ = hashMap1.Put("Removed", "deleted")
		_, _ = hashMap1.Remove("Removed")
		require.NoError(its.T(), client1.Sync())
	})
}
