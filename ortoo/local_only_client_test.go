package ortoo

import (
	"github.com/knowhunger/ortoo/ortoo/model"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLocalOnlyClientTest(t *testing.T) {
	client1, err := NewClient(NewLocalClientConfig("testCollection"), "localOnly1")
	require.NoError(t, err)
	client2, err := NewClient(NewLocalClientConfig("testCollection"), "localOnly2")
	require.NoError(t, err)
	intCounter1 := client1.CreateIntCounter("key", nil)
	_, _ = intCounter1.IncreaseBy(2)
	_, _ = intCounter1.IncreaseBy(3)
	meta, snapshot, err := intCounter1.(model.Datatype).GetMetaAndSnapshot()

	intCounter2 := client2.CreateIntCounter("key", nil)

	err = intCounter2.(model.Datatype).SetMetaAndSnapshot(meta, snapshot)
	require.NoError(t, err)
	require.Equal(t, intCounter1.Get(), intCounter2.Get())
}
