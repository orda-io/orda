package ortoo

import (
	"fmt"
	"github.com/knowhunger/ortoo/pkg/iface"
	"github.com/knowhunger/ortoo/pkg/log"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/pkg/testonly"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIntCounterTransactions(t *testing.T) {
	t.Run("Can transaction for Counter", func(t *testing.T) {
		tw := testonly.NewTestWire(true)

		counter1 := newCounter(testonly.NewBase("key1", model.TypeOfDatatype_COUNTER), tw, nil)

		require.NoError(t, counter1.DoTransaction("transaction1", func(counter CounterInTxn) error {
			_, _ = counter.IncreaseBy(2)
			require.Equal(t, int32(2), counter.Get())
			_, _ = counter.IncreaseBy(4)
			require.Equal(t, int32(6), counter.Get())
			return nil
		}))

		require.Equal(t, int32(6), counter1.Get())

		require.Error(t, counter1.DoTransaction("transaction2", func(intCounter CounterInTxn) error {
			_, _ = intCounter.IncreaseBy(3)
			require.Equal(t, int32(9), intCounter.Get())
			_, _ = intCounter.IncreaseBy(5)
			require.Equal(t, int32(14), intCounter.Get())
			return fmt.Errorf("err")
		}))
		require.Equal(t, int32(6), counter1.Get())
	})

	t.Run("Can sync Counter operations with Test wire", func(t *testing.T) {
		tw := testonly.NewTestWire(true)
		counter1 := newCounter(testonly.NewBase("key1", model.TypeOfDatatype_COUNTER), tw, nil)
		counter2 := newCounter(testonly.NewBase("key2", model.TypeOfDatatype_COUNTER), tw, nil)

		tw.SetDatatypes(counter1.(*counter).ManageableDatatype, counter2.(*counter).ManageableDatatype)

		i, oErr := counter1.Increase()
		require.NoError(t, oErr)

		require.Equal(t, i, int32(1))
		require.Equal(t, counter1.Get(), counter2.Get())

		err := counter1.DoTransaction("transaction1", func(intCounter CounterInTxn) error {
			_, _ = intCounter.IncreaseBy(-1)
			require.Equal(t, int32(0), intCounter.Get())
			_, _ = intCounter.IncreaseBy(-2)
			_, _ = intCounter.IncreaseBy(-3)
			require.Equal(t, int32(-5), intCounter.Get())
			return nil
		})
		require.NoError(t, err)

		log.Logger.Infof("%#v vs. %#v", counter1.Get(), counter2.Get())
		require.Equal(t, counter1.Get(), counter2.Get())

		err = counter1.DoTransaction("transaction2", func(intCounter CounterInTxn) error {
			_, _ = intCounter.IncreaseBy(1)
			_, _ = intCounter.IncreaseBy(2)
			require.Equal(t, int32(-2), intCounter.Get())
			_, _ = intCounter.IncreaseBy(3)
			_, _ = intCounter.IncreaseBy(4)
			return fmt.Errorf("fail to do the transaction")
		})
		require.Error(t, err)

		log.Logger.Infof("%#v vs. %#v", counter1.Get(), counter2.Get())
		require.Equal(t, counter1.Get(), counter2.Get())
	})

	t.Run("Can set and get counterSnapshot", func(t *testing.T) {
		counter1 := newCounter(testonly.NewBase("key1", model.TypeOfDatatype_COUNTER), nil, nil)
		_, _ = counter1.Increase()
		_, _ = counter1.IncreaseBy(1234)

		clone := newCounter(testonly.NewBase("key2", model.TypeOfDatatype_COUNTER), nil, nil)
		meta1, snap1, err := counter1.(iface.Datatype).GetMetaAndSnapshot()
		require.NoError(t, err)
		err = clone.(iface.Datatype).SetMetaAndSnapshot(meta1, snap1)
		require.NoError(t, err)
		meta2, snap2, err := clone.(iface.Datatype).GetMetaAndSnapshot()
		require.NoError(t, err)

		log.Logger.Infof("%v", string(snap1))
		log.Logger.Infof("%v", string(snap2))
		require.Equal(t, snap1, snap2)
		require.Equal(t, meta1, meta2)

		require.Equal(t, counter1.Get(), clone.Get())
	})
}
