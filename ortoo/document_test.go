package ortoo

import (
	"encoding/json"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/ortoo/model"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestJSONSnapshot(t *testing.T) {

	var strt1 = struct {
		K1 string
		K2 int
	}{
		K1: "hello",
		K2: 1234,
	}

	var arr1 = []interface{}{"world", 1234, 3.14}

	t.Run("Can put JSONElements to JSONObject and obtain the parent of children", func(t *testing.T) {
		opID1 := model.NewOperationID()
		jsonObj := newJSONObject(nil, model.OldestTimestamp)

		// { "K1": 1234, "K2": 3.14 }
		jsonObj.put("K1", 1234, opID1.Next().GetTimestamp())
		jsonObj.put("K2", 3.14, opID1.Next().GetTimestamp())

		// get the child "K1"
		je1 := jsonObj.getChildAsJSONElement("K1")
		log.Logger.Infof("%+v", je1.String())
		require.Equal(t, float64(1234), je1.getValue())

		// get parent of K1
		require.Equal(t, TypeJSONObject, je1.getParent().getType())
		parent := je1.getParentAsJSONObject()
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
		// unmarshaled := newJSONObject(nil, model.OldestTimestamp)
		// err = json.Unmarshal(m, unmarshaled)
		// require.NoError(t, err)
	})

	t.Run("Can put nested JSONObject to JSONObject", func(t *testing.T) {
		opID1 := model.NewOperationID()
		jsonObj := newJSONObject(nil, model.OldestTimestamp)

		jsonObj.put("K1", strt1, opID1.Next().GetTimestamp())
		log.Logger.Infof("%v", jsonObj)
		log.Logger.Infof("%v", marshal(t, jsonObj.GetAsJSONCompatible()))
		k1JSONObj := jsonObj.getChildAsJSONObject("K1")
		log.Logger.Infof("%v", k1JSONObj)
		log.Logger.Infof("%v", marshal(t, k1JSONObj.GetAsJSONCompatible()))

		// add struct ptr
		k1JSONObj.put("K3", &strt1, opID1.Next().GetTimestamp())
		log.Logger.Infof("%v", marshal(t, jsonObj.GetAsJSONCompatible()))

		parent := k1JSONObj.getParentAsJSONObject()
		require.Equal(t, jsonObj, parent)

		jsonObj.put("K2", arr1, opID1.Next().GetTimestamp())
		log.Logger.Infof("%v", jsonObj)
		log.Logger.Infof("%v", marshal(t, jsonObj.GetAsJSONCompatible()))

		jsonObj.put("K3", &arr1, opID1.Next().GetTimestamp())
		log.Logger.Infof("%v", jsonObj)
		log.Logger.Infof("%v", marshal(t, jsonObj.GetAsJSONCompatible()))

		// map is put in bundle.
		mp := make(map[string]interface{})
		mp["X"] = 1234
		mp["Y"] = []interface{}{"hi", strt1}
		jsonObj.put("K4", mp, opID1.Next().GetTimestamp())
		log.Logger.Infof("%v", jsonObj)
		log.Logger.Infof("%v", marshal(t, jsonObj.GetAsJSONCompatible()))

		require.Equal(t, "hello", jsonObj.getChildAsJSONObject("K1").getChildAsJSONElement("K1").getValue())
		// jsonObj.getAsJSONArray("K2").get(1)
		// require.Equal(t, "world", )

	})

	t.Run("Can marshal and unmarshal snapshots ", func(t *testing.T) {
		opID1 := model.NewOperationID()
		jsonObj := newJSONObject(nil, model.OldestTimestamp)
		jsonObj.put("A1", arr1, opID1.Next().GetTimestamp())

		jsonObj.put("K2", 3.14, opID1.Next().GetTimestamp())
		jsonObj.put("K3", "willBeDeleted", opID1.Next().GetTimestamp())
		jsonObj.DeleteCommonInObject(jsonObj.getTime(), "K3", opID1.Next().GetTimestamp())

		m, err := json.Marshal(jsonObj)
		require.NoError(t, err)
		log.Logger.Infof("%+v", string(m))
		clone := jsonObject{}
		err = json.Unmarshal(m, &clone)
		require.NoError(t, err)
		log.Logger.Infof("%+v", clone)
		// a1 := jsonObj.getChildAsJSONArray("A1")
		// a1.arrayDeleteLocal(1, 1, opID1.GetTimestamp())
		// log.Logger.Infof("%v", jsonObj.String())
		// a1.arrayInsertCommon(0, nil, opID1.Next().GetTimestamp(), strt1)
		// log.Logger.Infof("%v", jsonObj)

		// unmarshaled := newJSONObject(nil, model.OldestTimestamp)
		// err = json.Unmarshal(m, unmarshaled)
		// require.NoError(t, err)
	})

	t.Run("Can put JSONArray to JSONObject", func(t *testing.T) {
		opID1 := model.NewOperationID()
		jsonObj := newJSONObject(nil, model.OldestTimestamp)

		array := []interface{}{1234, 3.14, "hello"}
		jsonObj.put("K1", array, opID1.Next().GetTimestamp())
		log.Logger.Infof("%v", jsonObj.String())
		log.Logger.Infof("%v", marshal(t, jsonObj.GetAsJSONCompatible()))
		arr := jsonObj.getChildAsJSONArray("K1")
		log.Logger.Infof("%v", arr.String())
		val1, err := arr.getAsJSONElement(0)
		require.NoError(t, err)
		log.Logger.Infof("%v", val1)
		require.Equal(t, float64(1234), val1.getValue())

		// arr.arrayInsertCommon(0, opID1.Next().GetTimestamp(), "hi", "there")
		// log.Logger.Infof("%v", arr.String())
		// log.Logger.Infof("%v", marshal(t, arr.GetAsJSON()))

	})
}
