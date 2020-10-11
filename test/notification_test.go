package integration

import (
	"fmt"
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/log"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/pkg/ortoo"
	"github.com/stretchr/testify/require"
	"sync"
)

func (its *IntegrationTestSuite) TestNotification() {
	key := GetFunctionName()
	its.Run("Can notify remote change", func() {
		config := NewTestOrtooClientConfig(its.collectionName)
		client1 := ortoo.NewClient(config, "client1")

		err := client1.Connect()
		require.NoError(its.T(), err)
		defer func() {
			_ = client1.Close()
		}()

		client2 := ortoo.NewClient(config, "client2")
		require.NoError(its.T(), err)
		err = client2.Connect()
		require.NoError(its.T(), err)
		defer func() {
			_ = client2.Close()
		}()

		intCounter1 := client1.CreateCounter(key, nil)
		_, _ = intCounter1.Increase()
		require.NoError(its.T(), client1.Sync())

		fmt.Printf("Subscribed by client2\n")
		wg := sync.WaitGroup{}
		wg.Add(3)
		intCounter2 := client2.SubscribeCounter(key, ortoo.NewHandlers(
			func(dt ortoo.Datatype, old model.StateOfDatatype, new model.StateOfDatatype) {
				intCounter := dt.(ortoo.Counter)
				log.Logger.Infof("STATE: %s -> %s %d", old, new, intCounter.Get())
				require.Equal(its.T(), int32(1), intCounter.Get())
				wg.Done() // one time
			},
			func(dt ortoo.Datatype, opList []interface{}) {
				for _, op := range opList {
					log.Logger.Infof("OPERATION: %+v", op)
				}
				wg.Done() // two times
			},
			func(dt ortoo.Datatype, err ...errors.OrtooError) {
				require.NoError(its.T(), err[0])
			}))
		require.NoError(its.T(), client2.Sync())

		_, _ = intCounter1.IncreaseBy(10)
		require.NoError(its.T(), client1.Sync())
		wg.Wait()
		require.Equal(its.T(), intCounter1.Get(), intCounter2.Get())
	})
}
