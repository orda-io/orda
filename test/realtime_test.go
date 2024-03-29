package integration

import (
	"fmt"
	"github.com/orda-io/orda/client/pkg/errors"
	"github.com/orda-io/orda/client/pkg/log"
	"github.com/orda-io/orda/client/pkg/model"
	"github.com/orda-io/orda/client/pkg/orda"
	"sync"
	"time"

	"github.com/stretchr/testify/require"
)

func (its *IntegrationTestSuite) TestNotification() {
	key := GetFunctionName()

	its.Run("Can notify remote change", func() {
		config := NewTestOrdaClientConfig(its.collectionName, model.SyncType_REALTIME)
		client1 := orda.NewClient(config, "client1")
		require.NoError(its.T(), client1.Connect())
		defer func() {
			_ = client1.Close()
		}()

		client2 := orda.NewClient(config, "client2")
		require.NoError(its.T(), client2.Connect())
		defer func() {
			_ = client2.Close()
		}()

		intCounter1 := client1.CreateCounter(key, nil)
		_, _ = intCounter1.Increase()
		require.NoError(its.T(), client1.Sync())

		fmt.Printf("Subscribed by client2\n")
		opCount := 0
		wg1 := sync.WaitGroup{}
		wg1.Add(2)
		wg2 := sync.WaitGroup{}
		wg2.Add(1)
		intCounter2 := client2.SubscribeCounter(key, orda.NewHandlers(
			func(dt orda.Datatype, old model.StateOfDatatype, new model.StateOfDatatype) {
				intCounter := dt.(orda.Counter)
				log.Logger.Infof("STATE: %s -> %s %d", old, new, intCounter.Get())
				require.Equal(its.T(), int32(1), intCounter.Get())
				wg1.Done() // one time
			},
			func(dt orda.Datatype, opList []interface{}) {
				log.Logger.Infof("opList.size: %v", len(opList))
				for _, op := range opList {
					opCount++
					log.Logger.Infof("%d) OPERATION %+v", opCount, op)

				}

				if opCount == 2 {
					wg1.Done() // two times
				} else if opCount == 3 {
					wg2.Done()
				}
			},
			func(dt orda.Datatype, err ...errors.OrdaError) {
				require.NoError(its.T(), err[0])
			}))

		wg1.Wait()

		_, _ = intCounter1.IncreaseBy(10)
		wg2.Wait()
		require.Equal(its.T(), intCounter1.Get(), intCounter2.Get())
	})

	its.Run("Can test realtime delivery", func() {
		key := key + "-rt"
		config := NewTestOrdaClientConfig(its.collectionName, model.SyncType_REALTIME)
		client1 := orda.NewClient(config, "realtime_client1")
		require.NoError(its.T(), client1.Connect())
		defer func() {
			_ = client1.Close()
		}()

		client2 := orda.NewClient(config, "realtime_client2")
		require.NoError(its.T(), client2.Connect())
		defer func() {
			_ = client2.Close()
		}()
		wg1 := new(sync.WaitGroup)
		wg1.Add(1)
		wg3 := new(sync.WaitGroup)
		wg3.Add(1)
		counter1 := client1.CreateCounter(key, orda.NewHandlers(
			func(dt orda.Datatype, old model.StateOfDatatype, new model.StateOfDatatype) {
				if new == model.StateOfDatatype_SUBSCRIBED {
					wg1.Done()
					return
				}
				require.Fail(its.T(), "fail state")
			}, func(dt orda.Datatype, opList []interface{}) {
				for _, op := range opList {
					log.Logger.Infof("%v", op)
					wg3.Done()
				}
			}, nil))
		_, _ = counter1.Increase()

		require.True(its.T(), WaitTimeout(wg1, time.Second*5))

		wg2 := new(sync.WaitGroup)
		wg2.Add(3)
		log.Logger.Infof("SUBSCRIBED by client2")
		counter2 := client2.SubscribeCounter(key, orda.NewHandlers(
			func(dt orda.Datatype, old model.StateOfDatatype, new model.StateOfDatatype) {
				log.Logger.Infof("subscribe:%s -> %s", old, new)
				if new == model.StateOfDatatype_SUBSCRIBED {
					wg2.Done()
					return
				}
				require.Fail(its.T(), "fail state")
			},
			func(dt orda.Datatype, opList []interface{}) {
				for _, op := range opList {
					log.Logger.Infof("%v", op)
					wg2.Done()
				}
			}, nil))
		require.True(its.T(), WaitTimeout(wg2, time.Second*5))
		require.Equal(its.T(), int32(1), counter2.Get())

		_, _ = counter2.IncreaseBy(10)
		require.True(its.T(), WaitTimeout(wg3, time.Second*5))
		require.Equal(its.T(), counter1.Get(), counter2.Get())

	})
}
