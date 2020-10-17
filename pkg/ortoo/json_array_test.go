package ortoo

import (
	"github.com/knowhunger/ortoo/pkg/log"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/pkg/testonly"
	"github.com/knowhunger/ortoo/pkg/types"
	"github.com/stretchr/testify/require"
	"testing"
)

var strt1 = map[string]interface{}{
	"K1": "hello",
	"K2": float64(1234),
}

var arr1 = []interface{}{"world", float64(1234), 3.14}

func initJSONArrayAndInsertTest(t *testing.T, root *jsonObject, opID *model.OperationID) *jsonArray {
	// {"K1":["world",1234,3.14]}
	root.putCommon("K1", arr1, opID.Next().GetTimestamp())
	log.Logger.Infof("%v", testonly.Marshal(t, root.GetAsJSONCompatible()))

	array := root.getChildAsJSONArray("K1")
	require.Equal(t, root, array.getParent())
	require.Equal(t, opID.GetTimestamp(), array.getCreateTime())
	var err error

	// {"K1":["world",{ "K1": "hello", "K2": 1234 }, 1234,3.14]}
	_, _, err = root.InsertLocalInArray(array.getCreateTime(), 1, opID.Next().GetTimestamp(), strt1)
	require.NoError(t, err)
	log.Logger.Infof("%v", testonly.Marshal(t, root.GetAsJSONCompatible()))

	objChild := array.getJSONType(1)
	require.Equal(t, array, objChild.getParent())
	require.Equal(t, TypeJSONObject, objChild.getType())

	// {"K1":["world",{ "K1": "hello", "K2": 1234 }, ["world", 1234, 3.14], 1234,3.14]}
	_, _, err = root.InsertLocalInArray(array.getCreateTime(), 2, opID.Next().GetTimestamp(), arr1)
	require.NoError(t, err)
	log.Logger.Infof("%v", testonly.Marshal(t, root.GetAsJSONCompatible()))

	arrChild := array.getJSONType(2)
	require.Equal(t, array, arrChild.getParent())
	require.Equal(t, TypeJSONArray, arrChild.getType())

	return array
}

func TestJSONArray(t *testing.T) {

	t.Run("Can insert remotely in JSONArray", func(t *testing.T) {
		opID := model.NewOperationID()
		base := testonly.NewBase(t.Name(), model.TypeOfDatatype_DOCUMENT)
		root := newJSONObject(base, nil, model.OldestTimestamp())

		var arr = make([]interface{}, 0)
		_, err := root.PutCommonInObject(root.getCreateTime(), "K1", arr, opID.Next().GetTimestamp())
		require.NoError(t, err)
		_, err = root.PutCommonInObject(root.getCreateTime(), "K2", arr, opID.Next().GetTimestamp())
		require.NoError(t, err)
		log.Logger.Infof("%v", testonly.Marshal(t, root.GetAsJSONCompatible()))

		k1 := root.getAsJSONType("K1").(*jsonArray)
		k2 := root.getAsJSONType("K2").(*jsonArray)

		ts1 := opID.Next().GetTimestamp()  // older
		ts2 := opID.Next().GetTimestamp()  // newer
		values1 := []interface{}{"x", "y"} // newer
		values2 := []interface{}{"a", "b"} // older

		// insert into K1
		a1, err := root.InsertRemoteInArray(k1.getCreateTime(), model.OldestTimestamp(), ts2, values1...)
		require.NoError(t, err)
		require.Equal(t, a1, k1)
		a2, err := root.InsertRemoteInArray(k1.getCreateTime(), model.OldestTimestamp(), ts1, values2...)
		require.NoError(t, err)
		require.Equal(t, a2, k1)

		// insert into K2
		a3, err := root.InsertRemoteInArray(k2.getCreateTime(), model.OldestTimestamp(), ts1, values2...)
		require.NoError(t, err)
		require.Equal(t, a3, k2)
		a4, err := root.InsertRemoteInArray(k2.getCreateTime(), model.OldestTimestamp(), ts2, values1...)
		require.NoError(t, err)
		require.Equal(t, a4, k2)

		//
		x := k2.getJSONType(0)
		y := k2.getJSONType(1)
		require.NoError(t, err)
		require.Equal(t, k2, x.getParent(), y.getParent())
		require.NotEqual(t, x.getCreateTime(), y.getCreateTime())

		log.Logger.Infof("%v", testonly.Marshal(t, root.GetAsJSONCompatible()))
		require.Equal(t, k1.GetAsJSONCompatible(), k2.GetAsJSONCompatible())

		// marshal and unmarshal snapshot
		jsonObjectMarshalTest(t, root)
	})

	t.Run("Can delete something locally in JSONArray", func(t *testing.T) {
		opID := model.NewOperationID()
		base := testonly.NewBase(t.Name(), model.TypeOfDatatype_DOCUMENT)
		root := newJSONObject(base, nil, model.OldestTimestamp())
		var err error
		array := initJSONArrayAndInsertTest(t, root, opID)

		// delete a JSONElement
		// {"K1":[{ "K1": "hello", "K2": 1234 }, ["world", 1234, 3.14], 1234,3.14]}
		element1 := array.getJSONType(0)
		require.False(t, element1.isTomb())
		_, _, err = root.DeleteLocalInArray(array.getCreateTime(), 0, 1, opID.Next().GetTimestamp())
		require.NoError(t, err)
		require.True(t, element1.isTomb())
		log.Logger.Infof("%v", testonly.Marshal(t, root.GetAsJSONCompatible()))

		// delete a JSONObject
		// {"K1":[["world", 1234, 3.14], 1234,3.14]}
		element2 := array.getJSONType(0) // keep it in advance.
		require.False(t, element2.isTomb())
		_, _, err = root.DeleteLocalInArray(array.getCreateTime(), 0, 1, opID.Next().GetTimestamp())
		require.NoError(t, err)
		require.True(t, element2.isTomb())
		require.Equal(t, opID.GetTimestamp(), element2.getDeleteTime())
		log.Logger.Infof("%v", testonly.Marshal(t, root.GetAsJSONCompatible()))

		// delete a JSONArray
		// {"K1":[1234,3.14]}
		element3 := array.getJSONType(0) // keep it in advance.
		require.False(t, element3.isTomb())
		_, _, err = root.DeleteLocalInArray(array.getCreateTime(), 0, 1, opID.Next().GetTimestamp())
		require.NoError(t, err)
		require.True(t, element3.isTomb())
		require.Equal(t, opID.GetTimestamp(), element3.getDeleteTime())
		log.Logger.Infof("%v", testonly.Marshal(t, root.GetAsJSONCompatible()))

		// delete two values
		del1 := array.getJSONType(0)
		require.False(t, del1.isTomb())
		del2 := array.getJSONType(1)
		require.False(t, del2.isTomb())
		_, _, err = root.DeleteLocalInArray(array.getCreateTime(), 0, 2, opID.Next().GetTimestamp())
		require.NoError(t, err)
		require.Equal(t, 0, array.Size())
		ts := opID.GetTimestamp()
		require.Equal(t, ts.GetAndNextDelimiter(), del1.getDeleteTime())
		require.Equal(t, ts.GetAndNextDelimiter(), del2.getDeleteTime())
		log.Logger.Infof("%v", testonly.Marshal(t, root.GetAsJSONCompatible()))

		// marshal and unmarshal snapshot
		jsonObjectMarshalTest(t, root)
	})

	t.Run("Can delete something remotely in JSONArray", func(t *testing.T) {
		opID := model.NewOperationID()
		base := testonly.NewBase(t.Name(), model.TypeOfDatatype_DOCUMENT)
		root := newJSONObject(base, nil, model.OldestTimestamp())

		// init JSONObject
		// {"K1":["world",{ "K1": "hello", "K2": 1234 }, ["world", 1234, 3.14], 1234,3.14]}
		array := initJSONArrayAndInsertTest(t, root, opID)

		// delete NOT_EXISTING remotely
		tsNotExisting := model.NewTimestamp(0, 0, types.NewCUID(), 0)
		del1, errs := root.DeleteRemoteInArray(array.getCreateTime(), opID.Next().GetTimestamp(), []*model.Timestamp{tsNotExisting})
		require.Error(t, errs)
		require.Nil(t, del1)
		log.Logger.Infof("%v", testonly.Marshal(t, root.GetAsJSONCompatible()))

		// delete JSONElement, JSONObject, JSONArray
		je := array.getJSONType(0)
		jo := array.getJSONType(1)
		ja := array.getJSONType(2)

		// {"K1":[1234,3.14]}
		ts1 := opID.Next().GetTimestamp()

		del2, errs := root.DeleteRemoteInArray(array.getCreateTime(), ts1.Clone(), []*model.Timestamp{je.getCreateTime(), jo.getCreateTime(), ja.getCreateTime()})
		require.NoError(t, errs)
		require.Equal(t, 3, len(del2))
		require.True(t, je.isTomb())
		require.True(t, jo.isTomb())
		require.True(t, ja.isTomb())
		require.Equal(t, ts1.GetAndNextDelimiter(), je.getDeleteTime())
		require.Equal(t, ts1.GetAndNextDelimiter(), jo.getDeleteTime())
		require.Equal(t, ts1.GetAndNextDelimiter(), ja.getDeleteTime())
		log.Logger.Infof("%v", testonly.Marshal(t, root.GetAsJSONCompatible()))

		// if delete again with newer timestamp, then they should be deleted with newer timestamp.
		ts2 := opID.Next().GetTimestamp()
		ts2_2 := ts2.Clone()
		del3, errs := root.DeleteRemoteInArray(array.getCreateTime(), ts2.Clone(), []*model.Timestamp{je.getCreateTime(), jo.getCreateTime(), ja.getCreateTime()})
		require.NoError(t, errs)
		require.Equal(t, 0, len(del3))
		require.True(t, je.isTomb())
		require.True(t, jo.isTomb())
		require.True(t, ja.isTomb())
		require.Equal(t, ts2.GetAndNextDelimiter(), je.getDeleteTime())
		require.Equal(t, ts2.GetAndNextDelimiter(), jo.getDeleteTime())
		require.Equal(t, ts2.GetAndNextDelimiter(), ja.getDeleteTime())
		log.Logger.Infof("%v", testonly.Marshal(t, root.GetAsJSONCompatible()))

		// if delete again with older timestamp, then they should be deleted with newer timestamp.
		del4, errs := root.DeleteRemoteInArray(array.getCreateTime(), ts1.Clone(), []*model.Timestamp{je.getCreateTime(), jo.getCreateTime(), ja.getCreateTime()})
		require.NoError(t, errs)
		require.Equal(t, 0, len(del4))
		require.True(t, je.isTomb())
		require.True(t, jo.isTomb())
		require.True(t, ja.isTomb())
		require.Equal(t, ts2_2.GetAndNextDelimiter(), je.getDeleteTime())
		require.Equal(t, ts2_2.GetAndNextDelimiter(), jo.getDeleteTime())
		require.Equal(t, ts2_2.GetAndNextDelimiter(), ja.getDeleteTime())
		log.Logger.Infof("%v", testonly.Marshal(t, root.GetAsJSONCompatible()))

		// insert next to deleted one.
		arr5, err := root.InsertRemoteInArray(array.getCreateTime(), jo.getCreateTime(), opID.Next().GetTimestamp(), "E1")
		require.NoError(t, err)
		require.Equal(t, arr5, array)
		log.Logger.Infof("%v", testonly.Marshal(t, root.GetAsJSONCompatible()))

		// TODO: should test for updating a deleted one.
		// marshal and unmarshal snapshot
		jsonObjectMarshalTest(t, root)
	})

	t.Run("Can update values locally in JSONArray", func(t *testing.T) {
		opID := model.NewOperationID()
		base := testonly.NewBase(t.Name(), model.TypeOfDatatype_DOCUMENT)
		root := newJSONObject(base, nil, model.OldestTimestamp())

		// init JSONObject
		// {"K1":["world",{ "K1": "hello", "K2": 1234 }, ["world", 1234, 3.14], 1234,3.14]}
		array := initJSONArrayAndInsertTest(t, root, opID)

		oldOne1 := array.getJSONType(0)
		oldOne2 := array.getJSONType(1)
		oldOne3 := array.getJSONType(2)

		// update 3 nodes
		ts := opID.Next().GetTimestamp()
		targets, upd1, oErr := root.UpdateLocalInArray(array.getCreateTime(), 0, ts.Clone(), "updated1", "updated2", "updated3")
		require.NoError(t, oErr)
		require.Equal(t, 3, len(upd1))
		require.Equal(t, targets[0], oldOne1.getCreateTime())
		require.Equal(t, targets[1], oldOne2.getCreateTime())
		require.Equal(t, targets[2], oldOne3.getCreateTime())
		log.Logger.Infof("%v", testonly.Marshal(t, root.GetAsJSONCompatible()))
		newOne1 := array.getJSONType(0)
		newOne2 := array.getJSONType(1)
		newOne3 := array.getJSONType(2)

		// the old nodes should be tombstones.
		require.True(t, oldOne1.isTomb())
		require.True(t, oldOne2.isTomb())
		require.True(t, oldOne3.isTomb())
		// Except jsonElement(oldOne1), they should be in cemetery.
		require.Equal(t, 2, len(array.getCommon().cemetery))

		require.Equal(t, newOne1.getCreateTime(), oldOne1.getDeleteTime())
		require.Equal(t, newOne2.getCreateTime(), oldOne2.getDeleteTime())
		require.Equal(t, newOne3.getCreateTime(), oldOne3.getDeleteTime())

		// marshal and unmarshal snapshot
		jsonObjectMarshalTest(t, root)
	})

	t.Run("Can update value remotely in JSONArray", func(t *testing.T) {
		opID1 := model.NewOperationID()
		opID2 := model.NewOperationIDWithCUID(types.NewCUID())
		base := testonly.NewBase(t.Name(), model.TypeOfDatatype_DOCUMENT)
		root := newJSONObject(base, nil, model.OldestTimestamp())

		// init JSONObject
		// {"K1":["world",{ "K1": "hello", "K2": 1234 }, ["world", 1234, 3.14], 1234,3.14]}
		array := initJSONArrayAndInsertTest(t, root, opID1)

		oldOne1 := array.getJSONType(0) // "world"
		oldOne2 := array.getJSONType(1) // { "K1": "hello", "K2": 1234 }
		oldOne3 := array.getJSONType(2) // ["world", 1234, 3.14]

		// updates with older timestamp
		ts1 := opID2.GetTimestamp()
		upd1, errs := root.UpdateRemoteInArray(array.getCreateTime(), ts1.Clone(), []*model.Timestamp{oldOne1.getCreateTime()}, []interface{}{strt1})
		require.NoError(t, errs)
		require.NotEqual(t, upd1[0], oldOne1)
		require.False(t, oldOne1.isTomb())
		require.True(t, upd1[0].isTomb())
		require.Equal(t, 1, len(root.getCommon().cemetery))

		log.Logger.Infof("%v", testonly.Marshal(t, root.GetAsJSONCompatible()))

		// updates with newer timestamp
		ts2 := opID2.Next().GetTimestamp()
		upd2, errs := root.UpdateRemoteInArray(array.getCreateTime(), ts2.Clone(), []*model.Timestamp{oldOne1.getCreateTime()}, []interface{}{"updated1"})
		require.NoError(t, errs)
		require.Equal(t, oldOne1, upd2[0])
		require.True(t, oldOne1.isTomb())
		n0 := array.getJSONType(0)
		require.Equal(t, ts2, n0.getCreateTime())
		log.Logger.Infof("%v", testonly.Marshal(t, root.GetAsJSONCompatible()))

		// should find the first one with oldOne1.getCreateTime()
		ts3 := opID2.Next().GetTimestamp()
		upd3, errs := root.UpdateRemoteInArray(array.getCreateTime(), ts3.Clone(), []*model.Timestamp{oldOne1.getCreateTime()}, []interface{}{"updated2"})
		require.NoError(t, errs)
		require.Equal(t, upd3[0], n0)
		require.True(t, n0.isTomb())
		n1 := array.getJSONType(0)
		require.Equal(t, ts3, n1.getCreateTime())
		log.Logger.Infof("%v", testonly.Marshal(t, root.GetAsJSONCompatible()))

		// delete the first
		ts4 := opID2.Next().GetTimestamp()
		upd4, errs := root.DeleteRemoteInArray(array.getCreateTime(), ts4.Clone(), []*model.Timestamp{oldOne1.getCreateTime()})
		require.NoError(t, errs)
		require.Equal(t, upd4[0], n1)
		require.True(t, n1.isTomb())
		log.Logger.Infof("%v", testonly.Marshal(t, root.GetAsJSONCompatible()))

		// update the deleted one.
		ts5 := opID2.Next().GetTimestamp()
		_, errs = root.UpdateRemoteInArray(array.getCreateTime(), ts5.Clone(), []*model.Timestamp{oldOne1.getCreateTime()}, []interface{}{"updated3"})
		require.NoError(t, errs)
		// require.Equal(t, upd5, )
		log.Logger.Infof("%v", testonly.Marshal(t, root.GetAsJSONCompatible()))

		// update jsonObject and jsonArray at once.
		ts6 := opID2.Next().GetTimestamp()
		upd6, errs := root.UpdateRemoteInArray(array.getCreateTime(), ts6.Clone(),
			[]*model.Timestamp{oldOne2.getCreateTime(), oldOne3.getCreateTime()},
			[]interface{}{"updated4", "updated5"})
		require.NoError(t, errs)
		require.Equal(t, 2, len(upd6))
		require.Equal(t, upd6[0], oldOne2)
		require.Equal(t, upd6[1], oldOne3)
		require.True(t, oldOne2.isTomb())
		require.True(t, oldOne3.isTomb())
		require.Equal(t, ts6.GetAndNextDelimiter(), oldOne2.getDeleteTime())
		require.Equal(t, ts6.GetAndNextDelimiter(), oldOne3.getDeleteTime())

		log.Logger.Infof("%v", testonly.Marshal(t, root.GetAsJSONCompatible()))
		// marshal and unmarshal snapshot
		jsonObjectMarshalTest(t, root)
	})
}
