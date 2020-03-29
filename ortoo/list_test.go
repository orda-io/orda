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
		cuid1 := model.NewCUID()
		list1 := newList("key1", cuid1, tw, nil)

		list1.Insert(0, 1)

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
