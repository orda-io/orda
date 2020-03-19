package ortoo

import (
	"fmt"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/ortoo/model"
	"github.com/knowhunger/ortoo/ortoo/testonly"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIntCounterTransactions(t *testing.T) {
	tw := testonly.NewTestWire()
	cuid1 := model.NewCUID()

	intCounter1 := newIntCounter("key1", cuid1, tw, nil)

	require.NoError(t, intCounter1.DoTransaction("transaction1", func(intCounter IntCounterInTxn) error {
		_, _ = intCounter.IncreaseBy(2)
		require.Equal(t, int32(2), intCounter.Get())
		_, _ = intCounter.IncreaseBy(4)
		require.Equal(t, int32(6), intCounter.Get())
		return nil
	}))

	require.Equal(t, int32(6), intCounter1.Get())

	require.Error(t, intCounter1.DoTransaction("transaction2", func(intCounter IntCounterInTxn) error {
		_, _ = intCounter.IncreaseBy(3)
		require.Equal(t, int32(9), intCounter.Get())
		_, _ = intCounter.IncreaseBy(5)
		require.Equal(t, int32(14), intCounter.Get())
		return fmt.Errorf("err")
	}))
	require.Equal(t, int32(6), intCounter1.Get())
}

func TestIntCounterOperationSyncWithTestWire(t *testing.T) {
	tw := testonly.NewTestWire()
	intCounter1 := newIntCounter("key1", model.NewNilCUID(), tw, nil)
	intCounter2 := newIntCounter("key2", model.NewNilCUID(), tw, nil)

	tw.SetDatatypes(intCounter1, intCounter2)

	i, err := intCounter1.Increase()
	require.NoError(t, err)

	require.Equal(t, i, int32(1))
	require.Equal(t, intCounter1.Get(), intCounter2.Get())

	err = intCounter1.DoTransaction("transaction1", func(intCounter IntCounterInTxn) error {
		_, _ = intCounter.IncreaseBy(-1)
		require.Equal(t, int32(0), intCounter.Get())
		_, _ = intCounter.IncreaseBy(-2)
		_, _ = intCounter.IncreaseBy(-3)
		require.Equal(t, int32(-5), intCounter.Get())
		return nil
	})
	require.NoError(t, err)

	log.Logger.Printf("%#v vs. %#v", intCounter1.Get(), intCounter2.Get())
	require.Equal(t, intCounter1.Get(), intCounter2.Get())

	err = intCounter1.DoTransaction("transaction2", func(intCounter IntCounterInTxn) error {
		_, _ = intCounter.IncreaseBy(1)
		_, _ = intCounter.IncreaseBy(2)
		require.Equal(t, int32(-2), intCounter.Get())
		_, _ = intCounter.IncreaseBy(3)
		_, _ = intCounter.IncreaseBy(4)
		return fmt.Errorf("fail to do the transaction")
	})
	require.Error(t, err)

	log.Logger.Printf("%#v vs. %#v", intCounter1.Get(), intCounter2.Get())
	require.Equal(t, intCounter1.Get(), intCounter2.Get())
}
