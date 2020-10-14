package ortoo

import (
	"encoding/json"
	"fmt"
	"github.com/knowhunger/ortoo/pkg/log"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/pkg/testonly"
	"github.com/stretchr/testify/require"
	"testing"
)

func initList(t *testing.T, list *listSnapshot, opID *model.OperationID) {
	ts := opID.Next().GetTimestamp()
	target, _ := list.insertLocal(0, ts.Clone(), "x", "y")
	require.Equal(t, list.head.getOrderTime(), target)

	o1 := list.findOrderedType(0)
	o2 := list.findOrderedType(1)
	require.Equal(t, ts.GetAndNextDelimiter(), o1.getOrderTime())
	require.Equal(t, ts.GetAndNextDelimiter(), o2.getOrderTime())
	err := list.insertRemote(o2.getOrderTime(), ts.GetAndNextDelimiter(), "a", "b")
	require.NoError(t, err)
	log.Logger.Infof("%v", testonly.Marshal(t, list.GetAsJSONCompatible()))
}

func listIntegrityTest(t *testing.T, l *listSnapshot) {
	current := l.head
	for current != nil {
		next := current.getNext()
		if next != nil {
			require.Equal(t, current, next.getPrev())
		}
		current = current.getNext()
	}
}

func listMarshalTest(t *testing.T, original *listSnapshot) {
	clone := newListSnapshot(original.base)
	snap1, err2 := json.Marshal(original)
	require.NoError(t, err2)
	log.Logger.Infof("%v", string(snap1))
	err2 = json.Unmarshal(snap1, clone)
	require.NoError(t, err2)
	snap2, err2 := json.Marshal(clone)
	require.NoError(t, err2)
	log.Logger.Infof("%v", string(snap2))
	require.Equal(t, string(snap1), string(snap2))
	require.Equal(t, len(original.Map), len(clone.Map))

	e1 := original.head
	e2 := clone.head

	for e1 != nil && e2 != nil {
		require.Equal(t, e1.getValue(), e2.getValue())
		require.Equal(t, e1.getOrderTime(), e2.getOrderTime())
		require.Equal(t, e1.getTime(), e2.getTime())
		if e1.getPrev() != nil && e2.getPrev() != nil {
			require.Equal(t, e1.getPrev().getOrderTime(), e2.getPrev().getOrderTime())
		}
		if e1.getNext() != nil && e2.getNext() != nil {
			require.Equal(t, e1.getNext().getOrderTime(), e2.getNext().getOrderTime())
		}
		e1 = e1.getNext()
		e2 = e2.getNext()
	}
}

func TestList(t *testing.T) {

	t.Run("Can insert remotely in list", func(t *testing.T) {
		opID := model.NewOperationID()
		base := testonly.NewBase(t.Name(), model.TypeOfDatatype_LIST)
		list := newListSnapshot(base)

		oldTS1 := opID.Next().GetTimestamp()
		oldTS2 := opID.Next().GetTimestamp()
		err := list.insertRemote(model.OldestTimestamp(), oldTS2.Clone(), "x", "y")
		log.Logger.Infof("%v", testonly.Marshal(t, list.GetAsJSONCompatible()))
		require.NoError(t, err)
		err = list.insertRemote(model.OldestTimestamp(), oldTS1.Clone(), "a", "b")
		log.Logger.Infof("%v", testonly.Marshal(t, list.GetAsJSONCompatible()))
		require.NoError(t, err)
		n1 := list.findTimedType(0)

		require.Equal(t, "x", n1.getValue())
		require.Equal(t, oldTS2.GetAndNextDelimiter(), n1.getTime())

		n2 := list.findTimedType(1)
		require.Equal(t, "y", n2.getValue())
		require.Equal(t, oldTS2.GetAndNextDelimiter(), n2.getTime())

		n3 := list.findTimedType(2)
		require.Equal(t, "a", n3.getValue())
		require.Equal(t, oldTS1.GetAndNextDelimiter(), n3.getTime())

		n4 := list.findTimedType(3)
		require.Equal(t, "b", n4.getValue())
		require.Equal(t, oldTS1.GetAndNextDelimiter(), n4.getTime())

		listIntegrityTest(t, list)
		// marshal and unmarshal snapshot
		listMarshalTest(t, list)
	})

	t.Run("Can delete something in list", func(t *testing.T) {
		opID := model.NewOperationID()
		base := testonly.NewBase(t.Name(), model.TypeOfDatatype_LIST)
		list := newListSnapshot(base)

		// ["x","y","a","b"]
		initList(t, list, opID)

		// deleteLocal the first "x"
		e1 := list.findOrderedType(0)
		ts, _, v := list.deleteLocal(0, 1, opID.Next().GetTimestamp())
		require.Equal(t, "x", v[0])
		require.Equal(t, ts[0], e1.getOrderTime())
		require.Equal(t, opID.GetTimestamp(), e1.getTime())
		require.True(t, e1.isTomb())
		log.Logger.Infof("%v", testonly.Marshal(t, list.GetAsJSONCompatible()))

		// deleteRemote the first "x" again with an older timestamp
		deleted, errs := list.deleteRemote([]*model.Timestamp{e1.getOrderTime()}, model.OldestTimestamp())
		require.NoError(t, errs)
		require.Equal(t, 0, len(deleted))                          // nothing deleted effectively
		require.NotEqual(t, model.OldestTimestamp(), e1.getTime()) // this deletion is not effective.

		// deleteRemote the first "x" again with a newer timestamp
		deleted, errs = list.deleteRemote([]*model.Timestamp{e1.getOrderTime()}, opID.Next().GetTimestamp())
		require.NoError(t, errs)
		require.Equal(t, 0, len(deleted))                   // nothing deleted effectively
		require.Equal(t, opID.GetTimestamp(), e1.getTime()) // this deletion is effective.

		// list.updateRemote([]*model.Timestamp{e1.getOrderTime()}, []interface{}{"updated1"}, opID.Next().GetTimestamp())
		err := list.insertRemote(e1.getOrderTime(), opID.Next().GetTimestamp(), "X")
		require.NoError(t, err)
		require.Equal(t, 4, list.Size())
		log.Logger.Infof("%v", testonly.Marshal(t, list.GetAsJSONCompatible()))
		listIntegrityTest(t, list)
		// marshal and unmarshal snapshot
		listMarshalTest(t, list)
	})

	t.Run("Can update something in list", func(t *testing.T) {
		opID := model.NewOperationID()
		base := testonly.NewBase(t.Name(), model.TypeOfDatatype_LIST)
		list := newListSnapshot(base)

		// ["x","y","a","b"]
		initList(t, list, opID)

		// update locally
		updTS, updV, err := list.updateLocal(0, opID.Next().GetTimestamp(), []interface{}{"u1", "u2", "u3"})
		require.NoError(t, err)
		require.Equal(t, []interface{}{"x", "y", "a"}, updV)
		log.Logger.Infof("%v", testonly.Marshal(t, list.GetAsJSONCompatible()))

		e1 := list.findOrderedType(0)
		ts1 := e1.getTime()
		require.NotEqual(t, e1.getOrderTime(), e1.getTime())

		// update remotely with older timestamps
		upd, errs := list.updateRemote(updTS, []interface{}{"v1", "v2", "v3"}, model.OldestTimestamp())
		require.NoError(t, errs)
		require.Equal(t, 0, len(upd))
		require.Equal(t, ts1, e1.getTime())
		log.Logger.Infof("%v", testonly.Marshal(t, list.GetAsJSONCompatible()))

		// update remotely with newer timestamps
		upd, errs = list.updateRemote(updTS, []interface{}{"w1", "w2", "w3"}, opID.Next().GetTimestamp())
		require.NoError(t, errs)
		require.Equal(t, 3, len(upd))
		require.NotEqual(t, ts1, e1.getTime())

		log.Logger.Infof("%v", testonly.Marshal(t, list.GetAsJSONCompatible()))

		// delete with older timestamp; this should work
		dels, errs := list.deleteRemote([]*model.Timestamp{e1.getOrderTime()}, e1.getOrderTime().Clone())
		require.NoError(t, errs)
		require.True(t, dels[0].isTomb())
		log.Logger.Infof("%v", testonly.Marshal(t, list.GetAsJSONCompatible()))

		// update remotely with newer timestamps
		upd, errs = list.updateRemote(updTS, []interface{}{"x1", "x2", "x3"}, opID.Next().GetTimestamp())
		require.NoError(t, errs)
		require.Equal(t, 2, len(upd))
		require.Equal(t, e1.getOrderTime(), e1.getTime())
		require.True(t, e1.isTomb())
		log.Logger.Infof("%v", testonly.Marshal(t, list.GetAsJSONCompatible()))

		listIntegrityTest(t, list)
		// marshal and unmarshal snapshot
		listMarshalTest(t, list)
	})

	t.Run("Can perform list operations", func(t *testing.T) {
		tw := testonly.NewTestWire(false)
		list1 := newList(testonly.NewBase("key1", model.TypeOfDatatype_LIST), tw, nil)
		list2 := newList(testonly.NewBase("key2", model.TypeOfDatatype_LIST), tw, nil) // list2 always wins
		tw.SetDatatypes(list1.(*list).ManageableDatatype, list2.(*list).ManageableDatatype)

		// list1: x -> y
		inserted1, _ := list1.InsertMany(0, "x", "y")
		require.Equal(t, []interface{}{"x", "y"}, inserted1)
		json1 := testonly.Marshal(t, list1.GetAsJSON())
		require.Equal(t, `{"List":["x","y"]}`, json1)
		log.Logger.Infof("%s", json1)

		// list2: a -> b
		inserted2, _ := list2.InsertMany(0, "a", "b")
		require.Equal(t, []interface{}{"a", "b"}, inserted2)
		json2 := testonly.Marshal(t, list2.GetAsJSON())
		require.Equal(t, `{"List":["a","b"]}`, json2)
		log.Logger.Infof("%s", json2)

		tw.Sync()
		json3 := testonly.Marshal(t, list1.GetAsJSON())
		json4 := testonly.Marshal(t, list2.GetAsJSON())
		require.Equal(t, json3, json4)
		log.Logger.Infof("%s vs. %s", json3, json4)
		// log.Logger.Infof("SNAP1:%v", list1.(*list).snapshot)
		// log.Logger.Infof("SNAP2:%v", list2.(*list).snapshot)

		_, _ = list1.InsertMany(2, 7479)
		_, _ = list2.InsertMany(2, 3.141592)
		// log.Logger.Infof("SNAP1:%v", list1.(*list).snapshot)
		// log.Logger.Infof("SNAP2:%v", list2.(*list).snapshot)
		tw.Sync()
		json5 := testonly.Marshal(t, list1.GetAsJSON())
		json6 := testonly.Marshal(t, list2.GetAsJSON())
		log.Logger.Infof("SNAP1: %v => %v", json5, list1.(*list).snapshot)
		log.Logger.Infof("SNAP2: %v => %v", json6, list2.(*list).snapshot)
		require.Equal(t, json5, json6)

		updated1, _ := list1.Update(4, "X", "Y")
		require.Equal(t, []interface{}{"x", "y"}, updated1)
		updated2, _ := list2.Update(0, "A", "B")
		require.Equal(t, []interface{}{"a", "b"}, updated2)
		tw.Sync()
		json7 := testonly.Marshal(t, list1.GetAsJSON())
		json8 := testonly.Marshal(t, list2.GetAsJSON())
		require.Equal(t, json7, json8)
		log.Logger.Infof("SNAP1: %v => %v", json7, list1.(*list).snapshot)
		log.Logger.Infof("SNAP2: %v => %v", json8, list2.(*list).snapshot)

		m := make(map[string]string)
		m["a"] = "x"
		m["b"] = "y"
		time1 := "time.Now()" // TODO: should deal with time type
		_, _ = list1.Update(2, time1, m)
		_, _ = list2.Update(2, m, time1)
		log.Logger.Infof("%v", testonly.Marshal(t, list1.GetAsJSON()))
		log.Logger.Infof("%v", testonly.Marshal(t, list2.GetAsJSON()))
		tw.Sync()
		json9 := testonly.Marshal(t, list1.GetAsJSON())
		json10 := testonly.Marshal(t, list2.GetAsJSON())
		log.Logger.Infof("SNAP1: %v => %v", json9, list1.(*list).snapshot)
		log.Logger.Infof("SNAP2: %v => %v", json10, list2.(*list).snapshot)
		require.Equal(t, json9, json10)

		time2, err := list1.Get(3)
		require.NoError(t, err)
		require.Equal(t, time1, time2)

		deleted1, _ := list1.DeleteMany(0, 2)
		deleted2, _ := list2.DeleteMany(0, 2)
		require.Equal(t, 4, list1.Size())
		require.Equal(t, 4, list2.Size())
		log.Logger.Infof("%v vs %v", deleted1, deleted2)
		tw.Sync()
		require.Equal(t, 4, list1.Size())
		require.Equal(t, 4, list2.Size())
		json11 := testonly.Marshal(t, list1.GetAsJSON())
		json12 := testonly.Marshal(t, list2.GetAsJSON())
		log.Logger.Infof("SNAP1: %v => %v", json11, list1.(*list).snapshot)
		log.Logger.Infof("SNAP2: %v => %v", json12, list2.(*list).snapshot)

		_, err = list2.DeleteMany(0, 0)
		require.Error(t, err)

		marshaled, err2 := json.Marshal(list1.(*list).snapshot)
		require.NoError(t, err2)
		log.Logger.Infof("%v", string(marshaled))
		clone := listSnapshot{}
		err2 = json.Unmarshal(marshaled, &clone)
		require.NoError(t, err2)
		marshaled2, err2 := json.Marshal(&clone)
		require.NoError(t, err2)
		log.Logger.Infof("%v", string(marshaled2))
		log.Logger.Infof("%+v", list1.(*list).snapshot.GetAsJSONCompatible())
		log.Logger.Infof("%+v", clone.GetAsJSONCompatible())
	})

	t.Run("Can run transactions", func(t *testing.T) {
		tw := testonly.NewTestWire(true)
		list1 := newList(testonly.NewBase("key1", model.TypeOfDatatype_LIST), tw, nil)

		require.NoError(t, list1.DoTransaction("succeeded transaction", func(listTxn ListInTxn) error {
			_, _ = listTxn.InsertMany(0, "a", "b")
			gets1, err := listTxn.GetMany(0, 2)
			require.NoError(t, err)
			require.Equal(t, []interface{}{"a", "b"}, gets1)
			return nil
		}))
		gets1, err := list1.GetMany(0, 2)
		require.NoError(t, err)
		require.Equal(t, []interface{}{"a", "b"}, gets1)

		require.Error(t, list1.DoTransaction("failed transaction", func(listTxn ListInTxn) error {
			_, _ = listTxn.InsertMany(0, "x", "y")
			gets1, err := listTxn.GetMany(0, 2)
			require.NoError(t, err)
			require.Equal(t, []interface{}{"x", "y"}, gets1)
			return fmt.Errorf("fail")
		}))
		gets2, err := list1.GetMany(0, 2)
		require.NoError(t, err)
		require.Equal(t, []interface{}{"a", "b"}, gets2)
	})
}
