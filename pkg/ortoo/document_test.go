package ortoo

import (
	"encoding/json"
	"github.com/knowhunger/ortoo/pkg/errors"
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

func initJSONObjectForJSONArrayOperations(t *testing.T, root *jsonObject, opID *model.OperationID) *jsonArray {
	// {"K1":["world",1234,3.14]}
	root.putCommon("K1", arr1, opID.Next().GetTimestamp())
	log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))

	array := root.getChildAsJSONArray("K1")
	var err error

	// {"K1":["world",{ "K1": "hello", "K2": 1234 }, 1234,3.14]}
	_, _, err = root.InsertLocal(array.getKey(), 1, opID.Next().GetTimestamp(), strt1)
	require.NoError(t, err)
	log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))

	// {"K1":["world",{ "K1": "hello", "K2": 1234 }, ["world", 1234, 3.14], 1234,3.14]}
	_, _, err = root.InsertLocal(array.getKey(), 2, opID.Next().GetTimestamp(), arr1)
	require.NoError(t, err)
	log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))

	return array
}

func TestJSONSnapshot(t *testing.T) {

	t.Run("Can put JSONElements to JSONObject and obtain the parent of children", func(t *testing.T) {
		opID1 := model.NewOperationID()
		base := datatypes.NewBaseDatatype(t.Name(), model.TypeOfDatatype_DOCUMENT, types.NewCUID())
		jsonObj := newJSONObject(base, nil, model.OldestTimestamp)

		// { "K1": 1234, "K2": 3.14 }
		jsonObj.putCommon("K1", 1234, opID1.Next().GetTimestamp())
		jsonObj.putCommon("K2", 3.14, opID1.Next().GetTimestamp())

		// get the child "K1"
		je1 := jsonObj.getChildAsJSONElement("K1")
		log.Logger.Infof("%+v", je1.String())
		require.Equal(t, float64(1234), je1.getValue())

		// get parent of K1
		require.Equal(t, TypeJSONObject, je1.getParent().getType())
		parent := je1.getParent().(*jsonObject)
		require.Equal(t, jsonObj, parent)
		log.Logger.Infof("%+v", parent.String())

		// get the child "K2"
		je2 := parent.getChildAsJSONElement("K2")
		log.Logger.Infof("%+v", je2.String())
		require.Equal(t, 3.14, je2.getValue())

		log.Logger.Infof("%v", jsonObj.GetAsJSONCompatible())

		m, err := json.Marshal(jsonObj) // ==> jsonObject.MarshalJSON
		require.NoError(t, err)
		log.Logger.Infof("%s", string(m))
		unmarshaled := newJSONObject(base, nil, model.OldestTimestamp)
		err = json.Unmarshal(m, unmarshaled)
		require.NoError(t, err)
	})

	t.Run("Can delete something remotely in JSONObject", func(t *testing.T) {
		opID := model.NewOperationID()
		base := datatypes.NewBaseDatatype(t.Name(), model.TypeOfDatatype_DOCUMENT, types.NewCUID())
		root := newJSONObject(base, nil, model.OldestTimestamp)

		// { "K1": "hello" }
		root.putCommon("K1", "hello", opID.Next().GetTimestamp())
		// { "K1": "hello", "K2": { "K1": "hello", "K2": 1234 } }
		root.putCommon("K2", strt1, opID.Next().GetTimestamp())
		// { "K1": "hello", "K2": { "K1": "hello", "K2": 1234 }, "K3": ["world", 1234, 3.14] }
		root.putCommon("K3", arr1, opID.Next().GetTimestamp())

		// delete NOT_EXISTING remotely
		old0, err := root.DeleteRemoteInObject(root.getKey(), "NOT_EXISTING", opID.Next().GetTimestamp())
		require.Nil(t, old0)

		// delete a JSONElement remotely
		old1, err := root.DeleteRemoteInObject(root.getKey(), "K1", opID.Next().GetTimestamp())
		require.NoError(t, err)
		require.Equal(t, "hello", old1)

		// if deleting again, newer timestamp should remain
		old2, err := root.DeleteRemoteInObject(root.getKey(), "K1", opID.Next().GetTimestamp())
		require.NoError(t, err)
		require.Equal(t, "hello", old2)
		require.Equal(t, opID.GetTimestamp(), root.getChildAsJSONElement("K1").getTime())

		old3, err := root.DeleteRemoteInObject(root.getKey(), "K2", opID.Next().GetTimestamp())
		require.NoError(t, err)
		require.Equal(t, strt1, old3)
		log.Logger.Infof("%v", old3)

		old4, err := root.DeleteRemoteInObject(root.getKey(), "K3", opID.Next().GetTimestamp())
		require.NoError(t, err)
		log.Logger.Infof("%v", old4)
		require.Equal(t, arr1, old4)

		m, err2 := json.Marshal(root) // ==> jsonObject.MarshalJSON
		require.NoError(t, err2)
		unmarshaled := newJSONObject(base, nil, model.OldestTimestamp)
		err2 = json.Unmarshal(m, unmarshaled)
		require.NoError(t, err2)
		log.Logger.Infof("%v", marshal(t, unmarshaled.GetAsJSONCompatible()))
		require.Equal(t, marshal(t, root.GetAsJSONCompatible()), marshal(t, unmarshaled.GetAsJSONCompatible()))
	})

	t.Run("Can delete something remotely in JSONArray", func(t *testing.T) {
		opID := model.NewOperationID()
		base := datatypes.NewBaseDatatype(t.Name(), model.TypeOfDatatype_DOCUMENT, types.NewCUID())
		root := newJSONObject(base, nil, model.OldestTimestamp)

		// init JSONObject
		// {"K1":["world",{ "K1": "hello", "K2": 1234 }, ["world", 1234, 3.14], 1234,3.14]}
		array := initJSONObjectForJSONArrayOperations(t, root, opID)

		// delete NOT_EXISTING remotely
		tsNotExisting := &model.Timestamp{
			Era:       0,
			Lamport:   1,
			CUID:      types.NewCUID(),
			Delimiter: 0,
		}
		errs := root.DeleteRemoteInArray(array.getKey(), []*model.Timestamp{tsNotExisting}, opID.Next().GetTimestamp())
		require.Equal(t, 1, len(errs))
		log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))

		// delete a JSONElement, a JSONObject, a JSONArray
		je, err := array.getPrecedenceType(0)
		require.NoError(t, err)

		jo, err := array.getPrecedenceType(1)
		require.NoError(t, err)

		ja, err := array.getPrecededType(2)
		require.NoError(t, err)

		// {"K1":[1234,3.14]}
		ts1 := opID.Next().GetTimestamp()
		errs = root.DeleteRemoteInArray(array.getKey(), []*model.Timestamp{je.getKey(), jo.getKey(), ja.getKey()}, ts1)
		require.Equal(t, 0, len(errs))
		require.True(t, je.isTomb())
		require.True(t, jo.isTomb())
		require.True(t, ja.isTomb())
		require.Equal(t, ts1, je.getPrecedence())
		require.Equal(t, ts1, jo.getPrecedence())
		require.Equal(t, ts1, ja.getPrecedence())
		log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))
		log.Logger.Infof("%v %v %v", je.getPrecedence().ToString(), jo.getPrecedence().ToString(), ja.getPrecedence().ToString())

		// if delete again with newer timestamp, then they should be deleted with newer timestamp
		ts2 := opID.Next().GetTimestamp()
		errs = root.DeleteRemoteInArray(array.getKey(), []*model.Timestamp{je.getKey(), jo.getKey(), ja.getKey()}, ts2)
		require.Equal(t, 0, len(errs))
		require.True(t, je.isTomb())
		require.True(t, jo.isTomb())
		require.True(t, ja.isTomb())
		log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))
		log.Logger.Infof("%v %v %v", je.getPrecedence().ToString(), jo.getPrecedence().ToString(), ja.getPrecedence().ToString())
		require.Equal(t, ts2, je.getPrecedence())
		require.Equal(t, ts2, jo.getPrecedence())
		require.Equal(t, ts2, ja.getPrecedence())

	})

	t.Run("Can delete something locally in JSONObject", func(t *testing.T) {
		opID := model.NewOperationID()
		base := datatypes.NewBaseDatatype(t.Name(), model.TypeOfDatatype_DOCUMENT, types.NewCUID())
		root := newJSONObject(base, nil, model.OldestTimestamp)

		// { "K1": "hello" }
		root.putCommon("K1", "hello", opID.Next().GetTimestamp())
		// { "K1": "hello", "K2": { "K1": "hello", "K2": 1234 } }
		root.putCommon("K2", strt1, opID.Next().GetTimestamp())
		// { "K1": "hello", "K2": { "K1": "hello", "K2": 1234 }, "K3": ["world", 1234, 3.14] }
		root.putCommon("K3", arr1, opID.Next().GetTimestamp())

		log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))

		// delete not existing
		old0, err := root.DeleteLocalInObject(root.getKey(), "NOT_EXISTING", opID.Next().GetTimestamp())
		require.Equal(t, errors.ErrDatatypeNoOp.ToErrorCode(), errors.ToOrtooError(err).GetCode())
		require.Nil(t, old0)

		// delete an element normal
		old1, err := root.DeleteLocalInObject(root.getKey(), "K1", opID.Next().GetTimestamp())
		require.NoError(t, err)
		require.Equal(t, "hello", old1)

		// delete again: it should do nothing
		old2, err := root.DeleteLocalInObject(root.getKey(), "K1", opID.Next().GetTimestamp())
		require.Equal(t, errors.ErrDatatypeNoOp.ToErrorCode(), errors.ToOrtooError(err).GetCode())
		require.Nil(t, old2)
		log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))

		// delete a JSONObject
		old3, err := root.DeleteLocalInObject(root.getKey(), "K2", opID.Next().GetTimestamp())
		require.NoError(t, err)
		log.Logger.Infof("%v", old3)
		log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))

		// delete a JSONArray
		old4, err := root.DeleteLocalInObject(root.getKey(), "K3", opID.Next().GetTimestamp())
		require.NoError(t, err)
		log.Logger.Infof("%v", old4)
		log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))

		// marshal and unmarshal
		m, err2 := json.Marshal(root) // ==> jsonObject.MarshalJSON
		require.NoError(t, err2)
		unmarshaled := newJSONObject(base, nil, model.OldestTimestamp)
		err2 = json.Unmarshal(m, unmarshaled)
		require.NoError(t, err2)
		log.Logger.Infof("%v", marshal(t, unmarshaled.GetAsJSONCompatible()))
		require.Equal(t, marshal(t, root.GetAsJSONCompatible()), marshal(t, unmarshaled.GetAsJSONCompatible()))
	})

	t.Run("Can delete something locally in JSONArray", func(t *testing.T) {
		opID := model.NewOperationID()
		base := datatypes.NewBaseDatatype(t.Name(), model.TypeOfDatatype_DOCUMENT, types.NewCUID())
		root := newJSONObject(base, nil, model.OldestTimestamp)
		var err error
		array := initJSONObjectForJSONArrayOperations(t, root, opID)

		// delete out of index bound
		_, _, err = root.DeleteLocalInArray(array.getKey(), 10, 1, opID.Next().GetTimestamp())
		require.Error(t, err)
		log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))

		// delete a JSONElement
		// {"K1":[{ "K1": "hello", "K2": 1234 }, ["world", 1234, 3.14], 1234,3.14]}
		_, _, err = root.DeleteLocalInArray(array.getKey(), 0, 1, opID.Next().GetTimestamp())
		require.NoError(t, err)
		log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))

		// delete a JSONObject
		// {"K1":[["world", 1234, 3.14], 1234,3.14]}
		_, _, err = root.DeleteLocalInArray(array.getKey(), 0, 1, opID.Next().GetTimestamp())
		require.NoError(t, err)
		log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))

		// delete a JSONArray
		// {"K1":[1234,3.14]}
		_, _, err = root.DeleteLocalInArray(array.getKey(), 0, 1, opID.Next().GetTimestamp())
		require.NoError(t, err)
		log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))

		// delete two values
		_, _, err = root.DeleteLocalInArray(array.getKey(), 0, 2, opID.Next().GetTimestamp())
		require.NoError(t, err)
		log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))
		// marshal and unmarshal
		m, err2 := json.Marshal(root) // ==> jsonObject.MarshalJSON
		require.NoError(t, err2)
		unmarshaled := newJSONObject(base, nil, model.OldestTimestamp)
		err2 = json.Unmarshal(m, unmarshaled)
		require.NoError(t, err2)
		log.Logger.Infof("%v", marshal(t, unmarshaled.GetAsJSONCompatible()))
		require.Equal(t, marshal(t, root.GetAsJSONCompatible()), marshal(t, unmarshaled.GetAsJSONCompatible()))
	})

	t.Run("Can put nested JSONObject to JSONObject", func(t *testing.T) {
		opID1 := model.NewOperationID()
		base := datatypes.NewBaseDatatype(t.Name(), model.TypeOfDatatype_DOCUMENT, types.NewCUID())
		root := newJSONObject(base, nil, model.OldestTimestamp)

		// add struct
		root.putCommon("K1", strt1, opID1.Next().GetTimestamp())
		log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))

		// add struct ptr
		root.putCommon("K2", &strt1, opID1.Next().GetTimestamp())
		log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))

		k1 := root.getChildAsJSONObject("K1")
		k2 := root.getChildAsJSONObject("K2")
		require.Equal(t, marshal(t, k1.GetAsJSONCompatible()), marshal(t, k2.GetAsJSONCompatible()))

		// parent := k1.getParentAsJSONObject()
		// require.Equal(t, root, parent)
		//
		// root.putCommon("K2", arr1, opID1.Next().GetTimestamp())
		// log.Logger.Infof("%v", root)
		// log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))
		//
		// root.putCommon("K3", &arr1, opID1.Next().GetTimestamp())
		// log.Logger.Infof("%v", root)
		// log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))
		//
		// // map is putCommon in bundle.
		// mp := make(map[string]interface{})
		// mp["X"] = 1234
		// mp["Y"] = []interface{}{"hi", strt1}
		// root.putCommon("K4", mp, opID1.Next().GetTimestamp())
		// log.Logger.Infof("%v", root)
		// log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))
		//
		// require.Equal(t, "hello", root.getChildAsJSONObject("K1").getChildAsJSONElement("K1").getValue())
		// // root.getAsJSONArray("K2").get(1)
		// // require.Equal(t, "world", )

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
		jsonObj.DeleteRemoteInObject(jsonObj.getKey(), "K3", opID1.Next().GetTimestamp())

		// ["world", 1234, 3.14]
		a1 := jsonObj.getChildAsJSONArray("A1")
		// ["world", 3.14]
		_, _, _ = a1.arrayDeleteLocal(1, 1, opID1.GetTimestamp())
		log.Logger.Infof("%v", marshal(t, a1.GetAsJSONCompatible()))
		// [{"K1": "hello", "K2":" 1234}, "world", 3.14]
		_, _, _ = a1.arrayInsertCommon(0, nil, opID1.Next().GetTimestamp(), strt1)
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

	t.Run("Can put JSONArray to JSONObject", func(t *testing.T) {
		opID1 := model.NewOperationID()
		base := datatypes.NewBaseDatatype(t.Name(), model.TypeOfDatatype_DOCUMENT, types.NewCUID())
		jsonObj := newJSONObject(base, nil, model.OldestTimestamp)

		// put array for K1
		array := []interface{}{1234, 3.14, "hello"}
		jsonObj.putCommon("K1", array, opID1.Next().GetTimestamp())
		log.Logger.Infof("%v", jsonObj.String())
		log.Logger.Infof("%v", marshal(t, jsonObj.GetAsJSONCompatible()))

		// get jsonArray from jsonObject
		arr := jsonObj.getChildAsJSONArray("K1")
		log.Logger.Infof("%v", arr.String())

		// get jsonElement from jsonArray
		val1, err := arr.getPrecedenceType(0)
		require.NoError(t, err)
		log.Logger.Infof("%v", val1)
		require.Equal(t, float64(1234), val1.getValue())

		// insert to jsonArray
		arr.arrayInsertCommon(0, nil, opID1.Next().GetTimestamp(), "hi", "there")
		log.Logger.Infof("%v", arr.String())
		log.Logger.Infof("%v", marshal(t, arr.GetAsJSONCompatible()))
	})
}
