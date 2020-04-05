package ortoo

import (
	"fmt"
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
		m := make(map[string]string)
		m["a"] = "x"
		m["b"] = "y"
		time1 := time.Now()
		_, _ = list1.Update(2, time1, m)
		_, _ = list2.Update(2, m, time1)
		tw.Sync()
		json9, _ := list1.GetAsJSON()
		json10, _ := list2.GetAsJSON()
		log.Logger.Infof("SNAP1: %v => %v", json9, list1.(*list).snapshot)
		log.Logger.Infof("SNAP2: %v => %v", json10, list2.(*list).snapshot)
		require.Equal(t, json9, json10)
		time2, err := list1.Get(2)
		require.NoError(t, err)
		require.Equal(t, time1, time2.(time.Time))

		deleted1, _ := list1.DeleteMany(0, 2)
		deleted2, _ := list2.DeleteMany(0, 2)
		log.Logger.Infof("%v vs %v", deleted1, deleted2)
		tw.Sync()
		json11, _ := list1.GetAsJSON()
		json12, _ := list2.GetAsJSON()
		log.Logger.Infof("SNAP1: %v => %v", json11, list1.(*list).snapshot)
		log.Logger.Infof("SNAP2: %v => %v", json12, list2.(*list).snapshot)

		_, err = list2.DeleteMany(0, 0)
		require.Error(t, err)

	})

	t.Run("Can run transactions", func(t *testing.T) {
		tw := testonly.NewTestWire(true)
		cuid1 := model.NewCUID()
		list1 := newList("key1", cuid1, tw, nil)

		require.NoError(t, list1.DoTransaction("succeeded transaction", func(listTxn ListInTxn) error {
			_, _ = listTxn.Insert(0, "a", "b")
			gets1, err := listTxn.GetMany(0, 2)
			require.NoError(t, err)
			require.Equal(t, []interface{}{"a", "b"}, gets1)
			return nil
		}))
		gets1, err := list1.GetMany(0, 2)
		require.NoError(t, err)
		require.Equal(t, []interface{}{"a", "b"}, gets1)

		require.Error(t, list1.DoTransaction("failed transaction", func(listTxn ListInTxn) error {
			_, _ = listTxn.Insert(0, "x", "y")
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
