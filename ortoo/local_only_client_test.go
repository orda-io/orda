package ortoo

import (
	"encoding/json"
	"github.com/knowhunger/ortoo/ortoo/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLocalOnlyClientTest(t *testing.T) {
	t.Run("Can make local client", func(t *testing.T) {
		client1 := NewClient(NewLocalClientConfig("testCollection"), "localOnly1")
		client2 := NewClient(NewLocalClientConfig("testCollection"), "localOnly2")

		intCounter1 := client1.CreateIntCounter("key", nil)
		_, _ = intCounter1.IncreaseBy(2)
		_, _ = intCounter1.IncreaseBy(3)
		meta, snapshot, err := intCounter1.(types.Datatype).GetMetaAndSnapshot()
		require.NoError(t, err)
		intCounter2 := client2.CreateIntCounter("key", nil)
		snapB, err := json.Marshal(snapshot)
		err = intCounter2.(types.Datatype).SetMetaAndSnapshot(meta, string(snapB))
		require.NoError(t, err)
		require.Equal(t, intCounter1.Get(), intCounter2.Get())
	})

}
