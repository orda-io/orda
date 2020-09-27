package ortoo

import (
	"encoding/json"
	"github.com/knowhunger/ortoo/pkg/internal/datatypes"
	"github.com/knowhunger/ortoo/pkg/log"
	"github.com/knowhunger/ortoo/pkg/model"
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
	log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))

	array := root.getChildAsJSONArray("K1")
	require.Equal(t, root, array.getParent())
	require.Equal(t, opID.GetTimestamp(), array.getKeyTime())
	var err error

	// {"K1":["world",{ "K1": "hello", "K2": 1234 }, 1234,3.14]}
	_, _, err = root.InsertLocalInArray(array.getKeyTime(), 1, opID.Next().GetTimestamp(), strt1)
	require.NoError(t, err)
	log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))

	objChild, err := array.getJSONType(1)
	require.NoError(t, err)
	require.Equal(t, array, objChild.getParent())
	require.Equal(t, TypeJSONObject, objChild.getType())

	// {"K1":["world",{ "K1": "hello", "K2": 1234 }, ["world", 1234, 3.14], 1234,3.14]}
	_, _, err = root.InsertLocalInArray(array.getKeyTime(), 2, opID.Next().GetTimestamp(), arr1)
	require.NoError(t, err)
	log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))

	arrChild, err := array.getJSONType(2)
	require.NoError(t, err)
	require.Equal(t, array, arrChild.getParent())
	require.Equal(t, TypeJSONArray, arrChild.getType())

	return array
}

func TestJSONArray(t *testing.T) {

	t.Run("Can insert remotely in JSONArray", func(t *testing.T) {
		opID := model.NewOperationID()
		base := datatypes.NewBaseDatatype(t.Name(), model.TypeOfDatatype_DOCUMENT, types.NewCUID())
		root := newJSONObject(base, nil, model.OldestTimestamp)

		var arr = make([]interface{}, 0)
		_, err := root.PutCommonInObject(root.getKeyTime(), "K1", arr, opID.Next().GetTimestamp())
		require.NoError(t, err)
		_, err = root.PutCommonInObject(root.getKeyTime(), "K2", arr, opID.Next().GetTimestamp())
		require.NoError(t, err)
		log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))

		k1 := root.getAsJSONType("K1").(*jsonArray)
		k2 := root.getAsJSONType("K2").(*jsonArray)

		ts1 := opID.Next().GetTimestamp()  // older
		ts2 := opID.Next().GetTimestamp()  // newer
		values1 := []interface{}{"x", "y"} // newer
		values2 := []interface{}{"a", "b"} // older

		// insert into K1
		err = root.InsertRemoteInArray(k1.getKeyTime(), model.OldestTimestamp, ts2, values1...)
		require.NoError(t, err)
		err = root.InsertRemoteInArray(k1.getKeyTime(), model.OldestTimestamp, ts1, values2...)
		require.NoError(t, err)

		// insert into K2
		err = root.InsertRemoteInArray(k2.getKeyTime(), model.OldestTimestamp, ts1, values2...)
		require.NoError(t, err)
		err = root.InsertRemoteInArray(k2.getKeyTime(), model.OldestTimestamp, ts2, values1...)
		require.NoError(t, err)

		//
		x, err := k2.getJSONType(0)
		require.NoError(t, err)
		y, err := k2.getJSONType(1)
		require.NoError(t, err)
		require.Equal(t, k2, x.getParent(), y.getParent())
		require.NotEqual(t, x.getKeyTime(), y.getKeyTime())

		log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))
		require.Equal(t, k1.GetAsJSONCompatible(), k2.GetAsJSONCompatible())
	})

	t.Run("Can delete something locally in JSONArray", func(t *testing.T) {
		opID := model.NewOperationID()
		base := datatypes.NewBaseDatatype(t.Name(), model.TypeOfDatatype_DOCUMENT, types.NewCUID())
		root := newJSONObject(base, nil, model.OldestTimestamp)
		var err error
		array := initJSONArrayAndInsertTest(t, root, opID)

		// delete out of index bound
		_, _, err = root.DeleteLocalInArray(array.getKeyTime(), 10, 1, opID.Next().GetTimestamp())
		require.Error(t, err)
		log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))

		// delete a JSONElement
		// {"K1":[{ "K1": "hello", "K2": 1234 }, ["world", 1234, 3.14], 1234,3.14]}
		element1, err := array.getJSONType(0)
		require.NoError(t, err)
		require.False(t, element1.isTomb())
		_, _, err = root.DeleteLocalInArray(array.getKeyTime(), 0, 1, opID.Next().GetTimestamp())
		require.NoError(t, err)
		require.True(t, element1.isTomb())
		log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))

		// delete a JSONObject
		// {"K1":[["world", 1234, 3.14], 1234,3.14]}
		element2, err := array.getJSONType(0) // keep it in advance.
		require.NoError(t, err)
		require.False(t, element2.isTomb())
		_, _, err = root.DeleteLocalInArray(array.getKeyTime(), 0, 1, opID.Next().GetTimestamp())
		require.NoError(t, err)
		require.True(t, element2.isTomb())
		require.Equal(t, opID.GetTimestamp(), element2.getDelTime())
		log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))

		// delete a JSONArray
		// {"K1":[1234,3.14]}
		element3, err := array.getJSONType(0) // keep it in advance.
		require.NoError(t, err)
		require.False(t, element3.isTomb())
		_, _, err = root.DeleteLocalInArray(array.getKeyTime(), 0, 1, opID.Next().GetTimestamp())
		require.NoError(t, err)
		require.True(t, element3.isTomb())
		require.Equal(t, opID.GetTimestamp(), element3.getDelTime())
		log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))

		// delete two values
		del1, err := array.getJSONType(0)
		require.NoError(t, err)
		require.False(t, del1.isTomb())
		del2, err := array.getJSONType(1)
		require.NoError(t, err)
		require.False(t, del2.isTomb())
		_, _, err = root.DeleteLocalInArray(array.getKeyTime(), 0, 2, opID.Next().GetTimestamp())
		require.NoError(t, err)
		require.Equal(t, 0, array.Size())
		ts := opID.GetTimestamp()
		require.Equal(t, ts.GetAndNextDelimiter(), del1.getDelTime())
		require.Equal(t, ts.GetAndNextDelimiter(), del2.getDelTime())
		log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))

		// // marshal and unmarshal
		// m, err2 := json.Marshal(root) // ==> jsonObject.MarshalJSON
		// require.NoError(t, err2)
		// unmarshaled := newJSONObject(base, nil, model.OldestTimestamp)
		// err2 = json.Unmarshal(m, unmarshaled)
		// require.NoError(t, err2)
		// log.Logger.Infof("%v", marshal(t, unmarshaled.GetAsJSONCompatible()))
		// require.Equal(t, marshal(t, root.GetAsJSONCompatible()), marshal(t, unmarshaled.GetAsJSONCompatible()))
	})

	t.Run("Can delete something remotely in JSONArray", func(t *testing.T) {
		opID := model.NewOperationID()
		base := datatypes.NewBaseDatatype(t.Name(), model.TypeOfDatatype_DOCUMENT, types.NewCUID())
		root := newJSONObject(base, nil, model.OldestTimestamp)

		// init JSONObject
		// {"K1":["world",{ "K1": "hello", "K2": 1234 }, ["world", 1234, 3.14], 1234,3.14]}
		array := initJSONArrayAndInsertTest(t, root, opID)

		// delete NOT_EXISTING remotely
		tsNotExisting := model.NewTimestamp(0, 0, types.NewCUID(), 0)
		errs := root.DeleteRemoteInArray(array.getKeyTime(), opID.Next().GetTimestamp(), []*model.Timestamp{tsNotExisting})
		require.Equal(t, 1, len(errs))
		log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))

		// delete JSONElement, JSONObject, JSONArray
		je, err := array.getJSONType(0)
		require.NoError(t, err)
		jo, err := array.getJSONType(1)
		require.NoError(t, err)
		ja, err := array.getJSONType(2)
		require.NoError(t, err)

		// {"K1":[1234,3.14]}
		ts1 := opID.Next().GetTimestamp()

		errs = root.DeleteRemoteInArray(array.getKeyTime(), ts1.Clone(), []*model.Timestamp{je.getKeyTime(), jo.getKeyTime(), ja.getKeyTime()})
		require.Equal(t, 0, len(errs))
		require.True(t, je.isTomb())
		require.True(t, jo.isTomb())
		require.True(t, ja.isTomb())
		require.Equal(t, ts1.GetAndNextDelimiter(), je.getDelTime())
		require.Equal(t, ts1.GetAndNextDelimiter(), jo.getDelTime())
		require.Equal(t, ts1.GetAndNextDelimiter(), ja.getDelTime())
		log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))

		// if delete again with newer timestamp, then they should be deleted with newer timestamp.
		ts2 := opID.Next().GetTimestamp()
		ts2_2 := ts2.Clone()
		errs = root.DeleteRemoteInArray(array.getKeyTime(), ts2.Clone(), []*model.Timestamp{je.getKeyTime(), jo.getKeyTime(), ja.getKeyTime()})
		require.Equal(t, 0, len(errs))
		require.True(t, je.isTomb())
		require.True(t, jo.isTomb())
		require.True(t, ja.isTomb())
		require.Equal(t, ts2.GetAndNextDelimiter(), je.getDelTime())
		require.Equal(t, ts2.GetAndNextDelimiter(), jo.getDelTime())
		require.Equal(t, ts2.GetAndNextDelimiter(), ja.getDelTime())
		log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))

		// if delete again with older timestamp, then they should be deleted with newer timestamp.
		errs = root.DeleteRemoteInArray(array.getKeyTime(), ts1.Clone(), []*model.Timestamp{je.getKeyTime(), jo.getKeyTime(), ja.getKeyTime()})
		require.Equal(t, 0, len(errs))
		require.True(t, je.isTomb())
		require.True(t, jo.isTomb())
		require.True(t, ja.isTomb())
		require.Equal(t, ts2_2.GetAndNextDelimiter(), je.getDelTime())
		require.Equal(t, ts2_2.GetAndNextDelimiter(), jo.getDelTime())
		require.Equal(t, ts2_2.GetAndNextDelimiter(), ja.getDelTime())
		log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))

		// insert next to deleted one.
		err = root.InsertRemoteInArray(array.getKeyTime(), jo.getKeyTime(), opID.Next().GetTimestamp(), "E1")
		require.NoError(t, err)
		log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))

		// TODO: should test for updating a deleted one.
	})

	t.Run("Can marshal and unmarshal snapshots ", func(t *testing.T) {
		opID1 := model.NewOperationID()
		base := datatypes.NewBaseDatatype(t.Name(), model.TypeOfDatatype_DOCUMENT, types.NewCUID())
		jsonObj := newJSONObject(base, nil, model.OldestTimestamp)

		// { "A1": ["world", 1234, 3.14] }
		jsonObj.putCommon("A1", arr1, opID1.Next().GetTimestamp())

		// { "A1": ["world", 1234, 3.14], "K2": 3.14, "K3": "willBeDeleted" }
		jsonObj.putCommon("K2", 3.14, opID1.Next().GetTimestamp())
		jsonObj.putCommon("K3", "willBeDeleted", opID1.Next().GetTimestamp())
		// { "A1": ["world", 1234, 3.14], "K2": 3.14, "K3": "willBeDeleted"(TOMBSTONE) }
		jsonObj.DeleteRemoteInObject(jsonObj.getKeyTime(), "K3", opID1.Next().GetTimestamp())

		// ["world", 1234, 3.14]
		a1 := jsonObj.getChildAsJSONArray("A1")
		// ["world", 3.14]
		_, _, _ = a1.deleteLocal(1, 1, opID1.GetTimestamp())
		log.Logger.Infof("%v", marshal(t, a1.GetAsJSONCompatible()))
		// [{"K1": "hello", "K2":" 1234}, "world", 3.14]
		_, _, _ = a1.insertCommon(0, nil, opID1.Next().GetTimestamp(), strt1)
		log.Logger.Infof("%v", marshal(t, a1.GetAsJSONCompatible()))

		// marshaling
		m, err := json.Marshal(jsonObj)
		require.NoError(t, err)
		log.Logger.Infof("%+v", string(m))
		// unmarshaling
		clone := jsonObject{}
		err = json.Unmarshal(m, &clone)
		require.NoError(t, err)
		log.Logger.Infof("%+v", marshal(t, clone.GetAsJSONCompatible()))

		ck3 := clone.getChildAsJSONElement("K3")
		require.True(t, ck3.isTomb())
		ca1 := clone.getChildAsJSONArray("A1")
		v314, err := ca1.get(2)
		require.NoError(t, err)
		require.Equal(t, 3.14, v314)
	})

	t.Run("Can update values locally in JSONArray", func(t *testing.T) {
		opID := model.NewOperationID()
		base := datatypes.NewBaseDatatype(t.Name(), model.TypeOfDatatype_DOCUMENT, types.NewCUID())
		root := newJSONObject(base, nil, model.OldestTimestamp)

		// init JSONObject
		// {"K1":["world",{ "K1": "hello", "K2": 1234 }, ["world", 1234, 3.14], 1234,3.14]}
		array := initJSONArrayAndInsertTest(t, root, opID)

		oldOne1, oErr := array.getJSONType(0)
		require.NoError(t, oErr)
		oldOne2, oErr := array.getJSONType(1)
		require.NoError(t, oErr)
		oldOne3, oErr := array.getJSONType(2)
		require.NoError(t, oErr)

		// update 3 nodes
		ts := opID.Next().GetTimestamp()
		_, oErr = root.UpdateLocalInArray(array.getKeyTime(), 0, ts.Clone(), "updated1", "updated2", "updated3")
		require.NoError(t, oErr)
		log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))
		newOne1, oErr := array.getJSONType(0)
		require.NoError(t, oErr)
		newOne2, oErr := array.getJSONType(1)
		require.NoError(t, oErr)
		newOne3, oErr := array.getJSONType(2)
		require.NoError(t, oErr)

		// the old nodes should be tombstones.
		require.True(t, oldOne1.isTomb())
		require.True(t, oldOne2.isTomb())
		require.True(t, oldOne3.isTomb())
		// Except jsonElement(oldOne1), they should be in cemetery.
		require.Equal(t, 2, len(array.getRoot().cemetery))

		require.Equal(t, newOne1.getKeyTime(), oldOne1.getDelTime())
		require.Equal(t, newOne2.getKeyTime(), oldOne2.getDelTime())
		require.Equal(t, newOne3.getKeyTime(), oldOne3.getDelTime())

	})

	t.Run("Can update value remotely in JSONArray", func(t *testing.T) {
		opID1 := model.NewOperationID()
		opID2 := model.NewOperationIDWithCUID(types.NewCUID())
		base := datatypes.NewBaseDatatype(t.Name(), model.TypeOfDatatype_DOCUMENT, types.NewCUID())
		root := newJSONObject(base, nil, model.OldestTimestamp)

		// init JSONObject
		// {"K1":["world",{ "K1": "hello", "K2": 1234 }, ["world", 1234, 3.14], 1234,3.14]}
		array := initJSONArrayAndInsertTest(t, root, opID1)

		oldOne1, oErr := array.getJSONType(0)
		require.NoError(t, oErr)
		oldOne2, oErr := array.getJSONType(1)
		require.NoError(t, oErr)
		oldOne3, oErr := array.getJSONType(2)
		require.NoError(t, oErr)

		// updates with older timestamp
		ts1 := opID2.GetTimestamp()
		errs := root.UpdateRemoteInArray(array.getKeyTime(), ts1.Clone(), []*model.Timestamp{oldOne1.getKeyTime()}, []interface{}{strt1})
		require.Equal(t, 0, len(errs))
		require.False(t, oldOne1.isTomb())
		require.Equal(t, 1, len(root.getRoot().cemetery))

		log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))

		// updates with newer timestamp
		ts2 := opID2.Next().GetTimestamp()
		errs = root.UpdateRemoteInArray(array.getKeyTime(), ts2.Clone(), []*model.Timestamp{oldOne1.getKeyTime()}, []interface{}{"updated1"})
		require.Equal(t, 0, len(errs))
		require.True(t, oldOne1.isTomb())
		n0, oErr := array.getJSONType(0)
		require.NoError(t, oErr)
		require.Equal(t, ts2, n0.getKeyTime())
		log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))

		// should find the first one with oldOne1.getKeyTime()
		ts3 := opID2.Next().GetTimestamp()
		errs = root.UpdateRemoteInArray(array.getKeyTime(), ts3.Clone(), []*model.Timestamp{oldOne1.getKeyTime()}, []interface{}{"updated2"})
		require.Equal(t, 0, len(errs))
		require.True(t, n0.isTomb())
		n1, oErr := array.getJSONType(0)
		require.NoError(t, oErr)
		require.Equal(t, ts3, n1.getKeyTime())
		log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))

		// delete the first
		ts4 := opID2.Next().GetTimestamp()
		errs = root.DeleteRemoteInArray(array.getKeyTime(), ts4.Clone(), []*model.Timestamp{oldOne1.getKeyTime()})
		require.Equal(t, 0, len(errs))
		require.True(t, n1.isTomb())
		log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))

		// update the deleted one.
		ts5 := opID2.Next().GetTimestamp()
		errs = root.UpdateRemoteInArray(array.getKeyTime(), ts5.Clone(), []*model.Timestamp{oldOne1.getKeyTime()}, []interface{}{"updated3"})
		require.Equal(t, 0, len(errs))
		log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))

		// update jsonObject and jsonArray at once.
		ts6 := opID2.Next().GetTimestamp()
		errs = root.UpdateRemoteInArray(array.getKeyTime(), ts6.Clone(),
			[]*model.Timestamp{oldOne2.getKeyTime(), oldOne3.getKeyTime()},
			[]interface{}{"updated4", "updated5"})
		require.Equal(t, 0, len(errs))
		require.True(t, oldOne2.isTomb())
		require.True(t, oldOne3.isTomb())
		require.Equal(t, ts6.GetAndNextDelimiter(), oldOne2.getDelTime())
		require.Equal(t, ts6.GetAndNextDelimiter(), oldOne3.getDelTime())

		log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))
	})
}
