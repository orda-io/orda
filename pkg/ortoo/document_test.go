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

func TestJSONSnapshot(t *testing.T) {
	//
	// var strt1 = struct {
	// 	K1 string
	// 	K2 int
	// }{
	// 	K1: "hello",
	// 	K2: 1234,
	// }
	var strt1 = map[string]interface{}{
		"K1": "hello",
		"K2": float64(1234),
	}

	var arr1 = []interface{}{"world", float64(1234), 3.14}

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

	t.Run("Can remove something remotely in JSONObject", func(t *testing.T) {
		opID := model.NewOperationID()
		base := datatypes.NewBaseDatatype(t.Name(), model.TypeOfDatatype_DOCUMENT, types.NewCUID())
		root := newJSONObject(base, nil, model.OldestTimestamp)

		// { "K1": "hello" }
		root.putCommon("K1", "hello", opID.Next().GetTimestamp())
		// { "K1": "hello", "K2": { "K1": "hello", "K2": 1234 } }
		root.putCommon("K2", strt1, opID.Next().GetTimestamp())
		// { "K1": "hello", "K2": { "K1": "hello", "K2": 1234 }, "K3": ["world", 1234, 3.14] }
		root.putCommon("K3", arr1, opID.Next().GetTimestamp())

		old0, err := root.DeleteRemoteInObject(root.getKey(), "NOT_EXISTING", opID.Next().GetTimestamp())
		require.Nil(t, old0)

		old1, err := root.DeleteRemoteInObject(root.getKey(), "K1", opID.Next().GetTimestamp())
		require.NoError(t, err)
		require.Equal(t, "hello", old1)

		// if remove again, newer timestamp should remain
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

	t.Run("Can remove something locally in JSONObject", func(t *testing.T) {
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

		old0, err := root.DeleteLocalInObject(root.getKey(), "NOT_EXISTING", opID.Next().GetTimestamp())
		require.Equal(t, errors.ErrDatatypeNoOp.ToErrorCode(), errors.ToOrtooError(err).GetCode())
		require.Nil(t, old0)

		old1, err := root.DeleteLocalInObject(root.getKey(), "K1", opID.Next().GetTimestamp())
		require.NoError(t, err)
		require.Equal(t, "hello", old1)

		// delete again: it should do nothing
		old2, err := root.DeleteLocalInObject(root.getKey(), "K1", opID.Next().GetTimestamp())
		require.Equal(t, errors.ErrDatatypeNoOp.ToErrorCode(), errors.ToOrtooError(err).GetCode())
		require.Nil(t, old2)
		log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))

		old3, err := root.DeleteLocalInObject(root.getKey(), "K2", opID.Next().GetTimestamp())
		require.NoError(t, err)
		log.Logger.Infof("%v", old3)
		log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))

		old4, err := root.DeleteLocalInObject(root.getKey(), "K3", opID.Next().GetTimestamp())
		require.NoError(t, err)
		log.Logger.Infof("%v", old4)
		log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))

		m, err2 := json.Marshal(root) // ==> jsonObject.MarshalJSON
		require.NoError(t, err2)
		unmarshaled := newJSONObject(base, nil, model.OldestTimestamp)
		err2 = json.Unmarshal(m, unmarshaled)
		require.NoError(t, err2)
		log.Logger.Infof("%v", marshal(t, unmarshaled.GetAsJSONCompatible()))
		require.Equal(t, marshal(t, root.GetAsJSONCompatible()), marshal(t, unmarshaled.GetAsJSONCompatible()))
	})

	t.Run("Can remove something locally in JSONArray", func(t *testing.T) {
		opID := model.NewOperationID()
		base := datatypes.NewBaseDatatype(t.Name(), model.TypeOfDatatype_DOCUMENT, types.NewCUID())
		root := newJSONObject(base, nil, model.OldestTimestamp)

		// {"K1":["world",1234,3.14]}
		root.putCommon("K1", arr1, opID.Next().GetTimestamp())
		log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))

		array := root.getChildAsJSONArray("K1")
		_, _, err := root.InsertLocal(array.getKey(), 0, opID.Next().GetTimestamp(), strt1)
		require.NoError(t, err)

		log.Logger.Infof("%v", marshal(t, root.GetAsJSONCompatible()))
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
		val1, err := arr.getAsJSONElement(0)
		require.NoError(t, err)
		log.Logger.Infof("%v", val1)
		require.Equal(t, float64(1234), val1.getValue())

		// insert to jsonArray
		arr.arrayInsertCommon(0, nil, opID1.Next().GetTimestamp(), "hi", "there")
		log.Logger.Infof("%v", arr.String())
		log.Logger.Infof("%v", marshal(t, arr.GetAsJSONCompatible()))
	})
}
