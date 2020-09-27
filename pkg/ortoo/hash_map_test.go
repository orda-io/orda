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
		opID2 := model.NewOperationID()
		opID2.Lamport++
		opID3 := model.NewOperationID()
		opID3.Era++
		base := datatypes.NewBaseDatatype("test", model.TypeOfDatatype_HASH_MAP, types.NewCUID())
		snap := newHashMapSnapshot(base)
		_, _ = snap.putCommon("key1", "value1-1", opID1.GetTimestamp())
		_, _ = snap.putCommon("key1", "value1-2", opID2.GetTimestamp())

		_, _ = snap.putCommon("key2", "value2-1", opID2.GetTimestamp())
		_, _ = snap.putCommon("key2", "value2-2", opID1.GetTimestamp())

		json1 := snap.GetAsJSONCompatible()
		log.Logger.Infof("%+v", json1)
		j1, err := json.Marshal(json1)
		require.NoError(t, err)
		require.Equal(t, `{"key1":"value1-2","key2":"value2-1"}`, string(j1))
		require.Equal(t, 2, snap.size())

		removed1, err := snap.removeRemote("key1", opID3.GetTimestamp())
		require.NoError(t, err)
		removed2, err := snap.removeRemote("key1", opID1.GetTimestamp()) // remove with older timestamp; no op
		require.Error(t, err)
		require.Equal(t, "value1-2", removed1)
		require.Nil(t, removed2)
		json2 := snap.GetAsJSONCompatible()
		log.Logger.Infof("%+v", json2)
		j2, err := json.Marshal(json2)
		require.NoError(t, err)
		require.Equal(t, `{"key2":"value2-1"}`, string(j2))

		// marshal and unmarshal snapshot
		snap1, err := json.Marshal(snap)
		require.NoError(t, err)
		log.Logger.Infof("%v", string(snap1))
		clone := newHashMapSnapshot(base)
		err = json.Unmarshal(snap1, clone)
		require.NoError(t, err)
		snap2, err := json.Marshal(clone)
		require.NoError(t, err)
		log.Logger.Infof("%v", string(snap2))
	})
}
