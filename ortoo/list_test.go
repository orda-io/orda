package ortoo

import (
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/ortoo/model"
	"github.com/knowhunger/ortoo/ortoo/testonly"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestList(t *testing.T) {

	t.Run("Can perform list operations", func(t *testing.T) {
		tw := testonly.NewTestWire()
		list1 := newList("key1", model.NewCUID(), tw, nil)
		list2 := newList("key2", model.NewCUID(), tw, nil)
		tw.SetDatatypes(list1.(*list).FinalDatatype, list2.(*list).FinalDatatype)

		_, _ = list1.Insert(0, "x", "y")
		log.Logger.Infof("%v", list1.(*list).snapshot)
		_, _ = list1.Insert(2, "A", "B")
		log.Logger.Infof("%v", list1.(*list).snapshot)

	})

	t.Run("Can do operations with listSnapshot", func(t *testing.T) {
		snap := newListSnapshot()
		ts1 := model.NewOperationIDWithCuid(model.NewCUID()).GetTimestamp()
		ts2 := model.NewOperationIDWithCuid(model.NewCUID()).GetTimestamp()
		_, _, _ = snap.insertLocal(0, ts1, "hello", "world")
		_, _ = snap.insertRemote(model.OldestTimestamp.Hash(), ts2, "hi", "there")
		log.Logger.Infof("%v", snap)

		ts1.Lamport++
		_, _, _ = snap.insertLocal(0, ts1, 3.1415)
		log.Logger.Infof("%v", snap)

		_, _ = snap.deleteLocal(0, ts1)
		log.Logger.Infof("%v", snap)

		ts1 = ts1.Next()
		_, _ = snap.deleteLocal(0, ts1)
		log.Logger.Infof("%v", snap)
		ts1 = ts1.Next()
		_, _, _ = snap.insertLocal(0, ts1, "x")
		log.Logger.Infof("%v", snap)
		require.Equal(t, nil, snap.findNthNode(0).V) // should be head
		require.Equal(t, "x", snap.findNthNode(1).V)
		require.Equal(t, "world", snap.findNthNode(2).V)
		require.Equal(t, "hi", snap.findNthNode(3).V)
		require.Equal(t, "there", snap.findNthNode(4).V)
		_, _ = snap.deleteLocal(3, ts1)
		log.Logger.Infof("%v", snap)
		_, err := snap.deleteLocal(4, ts1.Next())
		require.Error(t, err)
	})
}
