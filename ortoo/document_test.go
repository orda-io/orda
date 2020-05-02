package ortoo

import (
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
		jsonObj.put("key1", int64(1234), opID1.Next().GetTimestamp())
		jsonObj.put("key2", 3.14, opID1.Next().GetTimestamp())

		je1 := jsonObj.getAsJSONElement("key1")
		log.Logger.Infof("%+v", je1.String())
		require.Equal(t, int64(1234), je1.getValue())

		require.Equal(t, TypeJSONObject, je1.getParent().getType())
		parent := je1.getParentAsJSONObject()
		require.Equal(t, jsonObj, parent)
		log.Logger.Infof("%+v", parent.String())

		je2 := parent.getAsJSONElement("key2")
		log.Logger.Infof("%+v", je2.String())
		require.Equal(t, 3.14, je2.getValue())

		log.Logger.Infof("%v", jsonObj.GetAsJSON())
	})

	t.Run("Can put nested JSONObject to JSONObject", func(t *testing.T) {
		opID1 := model.NewOperationID()
		jsonObj := newJSONObject(nil, model.OldestTimestamp)

		jsonObj.put("K1", strt1, opID1.Next().GetTimestamp())
		log.Logger.Infof("%v", jsonObj)
		log.Logger.Infof("%v", marshal(t, jsonObj.GetAsJSON()))
		k1JSONObj := jsonObj.getAsJSONObject("K1")
		log.Logger.Infof("%v", k1JSONObj)
		log.Logger.Infof("%v", marshal(t, k1JSONObj.GetAsJSON()))

		// add struct ptr
		k1JSONObj.put("K3", &strt1, opID1.Next().GetTimestamp())
		log.Logger.Infof("%v", marshal(t, jsonObj.GetAsJSON()))

		parent := k1JSONObj.getParentAsJSONObject()
		require.Equal(t, jsonObj, parent)

		jsonObj.put("K2", arr1, opID1.Next().GetTimestamp())
		log.Logger.Infof("%v", jsonObj)
		log.Logger.Infof("%v", marshal(t, jsonObj.GetAsJSON()))

		jsonObj.put("K3", &arr1, opID1.Next().GetTimestamp())
		log.Logger.Infof("%v", jsonObj)
		log.Logger.Infof("%v", marshal(t, jsonObj.GetAsJSON()))

		// map is put in bundle.
		mp := make(map[string]interface{})
		mp["X"] = 1234
		mp["Y"] = []interface{}{"hi", strt1}
		jsonObj.put("K4", mp, opID1.Next().GetTimestamp())
		log.Logger.Infof("%v", jsonObj)
		log.Logger.Infof("%v", marshal(t, jsonObj.GetAsJSON()))

		require.Equal(t, "hello", jsonObj.getAsJSONObject("K1").getAsJSONElement("K1").getValue())
		// jsonObj.getAsJSONArray("K2").get(1)
		// require.Equal(t, "world", )

	})

	t.Run("Can put JSONArray to JSONObject", func(t *testing.T) {
		opID1 := model.NewOperationID()
		jsonObj := newJSONObject(nil, model.OldestTimestamp)

		array := []interface{}{1234, 3.14, "hello"}
		jsonObj.put("K1", array, opID1.Next().GetTimestamp())
		log.Logger.Infof("%v", jsonObj.String())
		log.Logger.Infof("%v", marshal(t, jsonObj.GetAsJSON()))
		arr := jsonObj.getAsJSONArray("K1")
		val1, err := arr.getAsJSONElement(0)
		require.NoError(t, err)
		log.Logger.Infof("%v", val1)
		require.Equal(t, int64(1234), val1.getValue())

		arr.insertLocal(0, opID1.Next().GetTimestamp(), "hi", "there")
		log.Logger.Infof("%v", arr.String())
		log.Logger.Infof("%v", marshal(t, arr.GetAsJSON()))

	})
}
