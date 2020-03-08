package ortoo

import (
	"fmt"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/ortoo/model"
	"github.com/knowhunger/ortoo/ortoo/testonly"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHashMapTransactions(t *testing.T) {
	tw := testonly.NewTestWire()
	cuid1, err := model.NewCUID()
	require.NoError(t, err)

	hashMap1, err := newHashMap("key1", cuid1, tw)
	require.NoError(t, err)

	key1 := "k1"
	key2 := "k2"

	require.NoError(t, hashMap1.DoTransaction("transaction1", func(hm HashMapInTxn) error {
		_, _ = hm.Put(key1, 2)
		require.Equal(t, "2", hm.Get(key1))
		oldVal, _ := hm.Put(key1, 3)
		require.Equal(t, "2", oldVal)
		require.Equal(t, "3", hm.Get(key1))
		return nil
	}))
	require.Equal(t, "3", hashMap1.Get(key1))

	require.Error(t, hashMap1.DoTransaction("transaction2", func(hm HashMapInTxn) error {
		oldVal, _ := hm.Remove(key1)
		require.Equal(t, "3", oldVal)
		require.Equal(t, nil, hm.Get(key1))
		_, _ = hm.Put(key2, 5)
		require.Equal(t, "5", hm.Get(key2))
		return fmt.Errorf("error")
	}))
	require.Equal(t, "3", hashMap1.Get(key1))
	require.Equal(t, nil, hashMap1.Get(key2))
}

func TestHashMapSnapshot(t *testing.T) {
	snap := newHashMapSnapshot()
	opID1 := model.NewOperationID()
	opID2 := model.NewOperationID()
	opID3 := model.NewOperationID()
	opID2.Lamport++
	opID3.Era++
	snap.putCommon("key1", "value1-1", opID1.GetTimestamp())
	snap.putCommon("key1", "value1-2", opID2.GetTimestamp())

	snap.putCommon("key2", "value2-1", opID2.GetTimestamp())
	snap.putCommon("key2", "value2-2", opID1.GetTimestamp())
	snap1, err := snap.GetTypeAny()
	require.NoError(t, err)
	log.Logger.Infof("%+v", string(snap1.Value))

	removed1 := snap.remove("key1", opID3.GetTimestamp())
	removed2 := snap.remove("key2", opID1.GetTimestamp())
	require.Equal(t, "value1-2", removed1)
	require.Nil(t, removed2)
	snap2, err := snap.GetTypeAny()
	require.NoError(t, err)
	log.Logger.Infof("%+v", string(snap2.Value))
}
