package ortoo

import (
	"encoding/json"
	"fmt"
	"github.com/knowhunger/ortoo/pkg/iface"
	"github.com/knowhunger/ortoo/pkg/internal/datatypes"
	"github.com/knowhunger/ortoo/pkg/log"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/pkg/testonly"
	"github.com/knowhunger/ortoo/pkg/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHashMap(t *testing.T) {

	t.Run("Can run transaction", func(t *testing.T) {
		tw := testonly.NewTestWire(true)
		cuid1 := types.NewCUID()
		hashMap1 := newHashMap("key1", cuid1, tw, nil)
		key1 := "k1"
		key2 := "k2"

		require.NoError(t, hashMap1.DoTransaction("transaction success", func(hm HashMapInTxn) error {
			_, _ = hm.Put(key1, 2)
			require.Equal(t, float64(2), hm.Get(key1))
			oldVal, _ := hm.Put(key1, 3)
			require.Equal(t, float64(2), oldVal)
			require.Equal(t, float64(3), hm.Get(key1))
			return nil
		}))
		require.Equal(t, float64(3), hashMap1.Get(key1))

		require.Error(t, hashMap1.DoTransaction("transaction failure", func(hm HashMapInTxn) error {
			oldVal, _ := hm.Remove(key1)
			require.Equal(t, float64(3), oldVal)
			require.Equal(t, nil, hm.Get(key1))
			_, _ = hm.Put(key2, 5)
			require.Equal(t, float64(5), hm.Get(key2))
			return fmt.Errorf("fail")
		}))
		require.Equal(t, float64(3), hashMap1.Get(key1))
		require.Equal(t, nil, hashMap1.Get(key2))

		m, err := json.Marshal(hashMap1.(*hashMap).snapshot)
		require.NoError(t, err)
		log.Logger.Infof("%v", string(m))
		clone := hashMapSnapshot{}
		err = json.Unmarshal(m, &clone)
		require.NoError(t, err)
		m2, err := json.Marshal(hashMap1.(*hashMap).snapshot)
		require.NoError(t, err)
		log.Logger.Infof("%v", string(m2))
		require.Equal(t, m, m2)
	})

	t.Run("Can set and get snapshot", func(t *testing.T) {
		hashMap1 := newHashMap("key1", types.NewCUID(), nil, nil)
		hashMap1.Put("k1", 1)
		hashMap1.Put("k2", "2")
		hashMap1.Put("k3", 3.141592)
		hashMap1.Remove("k2")

		clone := newHashMap("key2", types.NewCUID(), nil, nil)
		meta1, snap1, err := hashMap1.(iface.Datatype).GetMetaAndSnapshot()
		require.NoError(t, err)
		snapA, err2 := json.Marshal(snap1)
		require.NoError(t, err2)
		err = clone.(iface.Datatype).SetMetaAndSnapshot(meta1, string(snapA))
		require.NoError(t, err)
		_, snap2, err := clone.(iface.Datatype).GetMetaAndSnapshot()
		require.NoError(t, err)
		snapB, err2 := json.Marshal(snap2)
		require.NoError(t, err)
		require.Equal(t, snapA, snapB)

		log.Logger.Infof("%v", string(snapA))
		log.Logger.Infof("%v", string(snapB))
	})

	t.Run("Can do operations with hashMapSnapshot", func(t *testing.T) {

		opID1 := model.NewOperationID()
		opID2 := model.NewOperationID().Next()

		base := datatypes.NewBaseDatatype("test", model.TypeOfDatatype_HASH_MAP, types.NewCUID())
		snap := newHashMapSnapshot(base)
		old1, err := snap.putCommon("key1", "v1", opID1.Next().GetTimestamp())
		require.NoError(t, err)
		require.Nil(t, old1)
		old2, err := snap.putCommon("key1", "v2", opID2.Next().GetTimestamp()) // should win
		require.NoError(t, err)
		require.Equal(t, "v1", old2)

		old3, err := snap.putCommon("key2", "v3", opID2.Next().GetTimestamp()) // should win
		require.NoError(t, err)
		require.Nil(t, old3)
		old4, err := snap.putCommon("key2", "v4", opID1.Next().GetTimestamp())
		require.NoError(t, err)
		require.Equal(t, "v4", old4)

		log.Logger.Infof("%+v", marshal(t, snap.GetAsJSONCompatible()))
		require.Equal(t, `{"key1":"v2","key2":"v3"}`, marshal(t, snap.GetAsJSONCompatible()))
		require.Equal(t, 2, snap.size())

		removed1, err := snap.removeRemote("key1", opID2.Next().GetTimestamp())
		require.NoError(t, err)
		require.Equal(t, "v2", removed1)

		removed2, err := snap.removeRemote("key2", model.OldestTimestamp()) // remove with older timestamp; not effective
		require.NoError(t, err)
		require.Nil(t, removed2)
		log.Logger.Infof("%+v", marshal(t, snap.GetAsJSONCompatible()))

		// marshal and unmarshal snapshot
		snap1, err2 := json.Marshal(snap)
		require.NoError(t, err2)
		log.Logger.Infof("%v", string(snap1))
		clone := newHashMapSnapshot(base)
		err2 = json.Unmarshal(snap1, clone)
		require.NoError(t, err2)
		snap2, err2 := json.Marshal(clone)
		require.NoError(t, err2)
		log.Logger.Infof("%v", string(snap2))
		require.Equal(t, string(snap1), string(snap2))
		require.Nil(t, clone.getFromMap("key1").getValue())
	})
}
