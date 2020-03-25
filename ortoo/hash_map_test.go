package ortoo

import (
	"fmt"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/ortoo/model"
	"github.com/knowhunger/ortoo/ortoo/testonly"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHashMap(t *testing.T) {

	t.Run("Can run transaction", func(t *testing.T) {
		tw := testonly.NewTestWire()
		cuid1 := model.NewCUID()
		hashMap1 := newHashMap("key1", cuid1, tw, nil)
		key1 := "k1"
		key2 := "k2"

		require.NoError(t, hashMap1.DoTransaction("transaction1", func(hm HashMapInTxn) error {
			_, _ = hm.Put(key1, 2)
			require.Equal(t, int64(2), hm.Get(key1))
			oldVal, _ := hm.Put(key1, 3)
			require.Equal(t, int64(2), oldVal)
			require.Equal(t, int64(3), hm.Get(key1))
			return nil
		}))
		require.Equal(t, int64(3), hashMap1.Get(key1))

		require.Error(t, hashMap1.DoTransaction("transaction2", func(hm HashMapInTxn) error {
			oldVal, _ := hm.Remove(key1)
			require.Equal(t, int64(3), oldVal)
			require.Equal(t, nil, hm.Get(key1))
			_, _ = hm.Put(key2, 5)
			require.Equal(t, int64(5), hm.Get(key2))
			return fmt.Errorf("error")
		}))
		require.Equal(t, int64(3), hashMap1.Get(key1))
		require.Equal(t, nil, hashMap1.Get(key2))
	})

	t.Run("Can set and get snapshot", func(t *testing.T) {
		hashMap1 := newHashMap("key1", model.NewCUID(), nil, nil)
		hashMap1.Put("k1", 1)
		hashMap1.Put("k2", "2")
		hashMap1.Put("k3", 3.141592)
		hashMap1.Remove("k2")

		clone := newHashMap("key2", model.NewCUID(), nil, nil)
		meta, snap, err := hashMap1.(model.Datatype).GetMetaAndSnapshot()
		require.NoError(t, err)
		err = clone.(model.Datatype).SetMetaAndSnapshot(meta, snap)
		require.NoError(t, err)
	})

	t.Run("Can test hashMapSnapshot", func(t *testing.T) {
		snap := newHashMapSnapshot()
		opID1 := model.NewOperationID()
		opID2 := model.NewOperationID()
		opID3 := model.NewOperationID()
		opID2.Lamport++
		opID3.Era++
		_, _ = snap.putCommon("key1", "value1-1", opID1.GetTimestamp())
		_, _ = snap.putCommon("key1", "value1-2", opID2.GetTimestamp())

		_, _ = snap.putCommon("key2", "value2-1", opID2.GetTimestamp())
		_, _ = snap.putCommon("key2", "value2-2", opID1.GetTimestamp())
		snap1, err := snap.GetAsJSON()
		require.NoError(t, err)
		log.Logger.Infof("%+v", snap1)
		require.Equal(t, `{"key1":"value1-2","key2":"value2-1"}`, snap1)

		removed1 := snap.removeCommon("key1", opID3.GetTimestamp())
		removed2 := snap.removeCommon("key2", opID1.GetTimestamp())
		require.Equal(t, "value1-2", removed1)
		require.Nil(t, removed2)
		snap2, err := snap.GetAsJSON()
		require.NoError(t, err)
		log.Logger.Infof("%+v", snap2)
		require.Equal(t, `{"key2":"value2-1"}`, snap2)
	})
}
