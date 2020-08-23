package ortoo

import (
	"encoding/json"
	"github.com/knowhunger/ortoo/ortoo/iface"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLocalOnlyClientTest(t *testing.T) {
	t.Run("Can make local client", func(t *testing.T) {
		client1 := NewClient(NewLocalClientConfig("testCollection"), "localOnly1")
		client2 := NewClient(NewLocalClientConfig("testCollection"), "localOnly2")

		intCounter1 := client1.CreateCounter("key", nil)
		_, _ = intCounter1.IncreaseBy(2)
		_, _ = intCounter1.IncreaseBy(3)
		meta, snapshot, err := intCounter1.(iface.Datatype).GetMetaAndSnapshot()
		require.NoError(t, err)
		intCounter2 := client2.CreateCounter("key", nil)
		snapB, err2 := json.Marshal(snapshot)
		require.NoError(t, err2)
		err = intCounter2.(iface.Datatype).SetMetaAndSnapshot(meta, string(snapB))
		require.NoError(t, err)
		require.Equal(t, intCounter1.Get(), intCounter2.Get())
	})

}
