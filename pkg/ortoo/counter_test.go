package ortoo

import (
	"fmt"
	"github.com/knowhunger/ortoo/pkg/log"
	"github.com/knowhunger/ortoo/pkg/testonly"
	"github.com/knowhunger/ortoo/pkg/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIntCounterTransactions(t *testing.T) {
	t.Run("Can transaction for Counter", func(t *testing.T) {
		tw := testonly.NewTestWire(true)
		cuid1 := types.NewCUID()

		counter1 := newCounter("key1", cuid1, tw, nil)

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
		counter1 := newCounter("key1", types.NewNilCUID(), tw, nil)
		counter2 := newCounter("key2", types.NewNilCUID(), tw, nil)

		tw.SetDatatypes(counter1.(*counter).ManageableDatatype, counter2.(*counter).ManageableDatatype)

		i, err := counter1.Increase()
		require.NoError(t, err)

		require.Equal(t, i, int32(1))
		require.Equal(t, counter1.Get(), counter2.Get())

		err = counter1.DoTransaction("transaction1", func(intCounter CounterInTxn) error {
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
}
