package ortoo

import (
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/ortoo/model"
	"github.com/knowhunger/ortoo/ortoo/testonly"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestList(t *testing.T) {

	t.Run("Can perform list operations", func(t *testing.T) {
		tw := testonly.NewTestWire(false)
		list1 := newList("key1", model.NewCUID(), tw, nil)
		list2 := newList("key2", model.NewCUID(), tw, nil)
		tw.SetDatatypes(list1.(*list).FinalDatatype, list2.(*list).FinalDatatype)

		inserted1, _ := list1.Insert(0, "x", "y")
		require.Equal(t, []interface{}{"x", "y"}, inserted1)
		json1, err := list1.GetAsJSON()
		require.NoError(t, err)
		require.Equal(t, `["x","y"]`, json1)
		log.Logger.Infof("%s", json1)

		inserted2, _ := list2.Insert(0, "a", "b")
		require.Equal(t, []interface{}{"a", "b"}, inserted2)
		json2, err := list2.GetAsJSON()
		require.NoError(t, err)
		require.Equal(t, `["a","b"]`, json2)
		log.Logger.Infof("%s", json2)

		tw.Sync()
		json3, err := list1.GetAsJSON()
		require.NoError(t, err)
		json4, err := list2.GetAsJSON()
		require.NoError(t, err)
		require.Equal(t, json3, json4)
		log.Logger.Infof("%s vs. %s", json3, json4)
		log.Logger.Infof("SNAP1:%v", list1.(*list).snapshot)
		log.Logger.Infof("SNAP2:%v", list2.(*list).snapshot)

		_, _ = list1.Insert(2, 7479)
		_, _ = list1.Insert(2, 3.141592)
		log.Logger.Infof("SNAP1:%v", list1.(*list).snapshot)
		log.Logger.Infof("SNAP2:%v", list2.(*list).snapshot)
		tw.Sync()
		json5, _ := list1.GetAsJSON()
		json6, _ := list2.GetAsJSON()
		log.Logger.Infof("SNAP1: %v => %v", json5, list1.(*list).snapshot)
		log.Logger.Infof("SNAP2: %v => %v", json6, list2.(*list).snapshot)
		require.Equal(t, json5, json6)

		updated1, _ := list1.Update(4, "X", "Y")
		require.Equal(t, []interface{}{"x", "y"}, updated1)
		updated2, _ := list2.Update(0, "A", "B")
		require.Equal(t, []interface{}{"a", "b"}, updated2)
		tw.Sync()
		json7, _ := list1.GetAsJSON()
		json8, _ := list2.GetAsJSON()
		require.Equal(t, json7, json8)
		log.Logger.Infof("SNAP1: %v => %v", json7, list1.(*list).snapshot)
		log.Logger.Infof("SNAP2: %v => %v", json8, list2.(*list).snapshot)
	})

}

func TestListSnapshot(t *testing.T) {
	t.Run("Can do operations with listSnapshot", func(t *testing.T) {
		snap := newListSnapshot()
		ts1 := model.NewOperationIDWithCuid(model.NewCUID()).GetTimestamp()
		ts2 := model.NewOperationIDWithCuid(model.NewCUID()).GetTimestamp()

		_, _, _ = snap.insertLocal(0, ts1, "hello", "world")
		require.Equal(t, int32(2), snap.size)
		n1, err := snap.get(1)
		require.NoError(t, err)
		require.Equal(t, "world", n1)

		_, _ = snap.insertRemote(model.OldestTimestamp.Hash(), ts2, "hi", "there")
		require.Equal(t, int32(4), snap.size)
		log.Logger.Infof("%v", snap)
		n2, err := snap.get(4)
		require.Error(t, err)
		require.Nil(t, n2)

		ts1.Lamport++
		_, _, _ = snap.insertLocal(0, ts1, 3.1415)
		log.Logger.Infof("%v", snap)

		_, deleted1, _ := snap.deleteLocal(0, 1, ts1)
		require.Equal(t, 3.1415, deleted1[0])
		log.Logger.Infof("%v", snap)

		ts1 = ts1.Next()
		_, deleted2, _ := snap.deleteLocal(0, 1, ts1)
		require.Equal(t, "hello", deleted2[0])
		log.Logger.Infof("%v", snap)
		ts1 = ts1.Next()
		_, _, _ = snap.insertLocal(0, ts1, "x")
		log.Logger.Infof("%v", snap)
		require.Equal(t, nil, snap.findNthTarget(0).V) // should be head
		require.Equal(t, "x", snap.findNthTarget(1).V)
		require.Equal(t, "world", snap.findNthTarget(2).V)
		require.Equal(t, "hi", snap.findNthTarget(3).V)
		require.Equal(t, "there", snap.findNthTarget(4).V)
		_, deleted3, _ := snap.deleteLocal(3, 1, ts1)
		log.Logger.Infof("%v", snap)
		require.Equal(t, "there", deleted3[0])
		require.Equal(t, int32(3), snap.size)
		ts1 = ts1.Next()
		_, _, err = snap.deleteLocal(4, 1, ts1)
		require.Error(t, err)
		ts1 = ts1.Next()
		ret, deleted4, err := snap.deleteLocal(0, 3, ts1)
		require.Equal(t, int32(0), snap.size)
		log.Logger.Infof("%v", snap)
		log.Logger.Infof("%v", ret)
		log.Logger.Infof("%v", deleted4)
		require.Equal(t, []interface{}{"x", "world", "hi"}, deleted4)

		ts1 = ts1.Next()
		ti := time.Now()
		_, _, err = snap.insertLocal(0, ts1, ti)
		require.NoError(t, err)
		log.Logger.Infof("%v", snap)
		g1, err := snap.get(0)
		require.Equal(t, ti, g1)

	})
}
