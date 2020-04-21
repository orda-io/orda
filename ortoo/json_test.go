package ortoo

import (
	"encoding/json"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/ortoo/model"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestJSONSnapshot(t *testing.T) {

	t.Run("Can put JSONElements to JSONObject and obtain the parent of children", func(t *testing.T) {
		opID1 := model.NewOperationID()
		jsonObj := newJSONObject(nil, model.OldestTimestamp)
		jsonObj.put("key1", int64(1234), opID1.Next().GetTimestamp())
		jsonObj.put("key2", 3.14, opID1.Next().GetTimestamp())

		je1 := jsonObj.getAsJSONElement("key1")
		log.Logger.Infof("%+v", je1.String())
		require.Equal(t, int64(1234), je1.getValue())

		require.Equal(t, typeJSONObject, je1.getParent().getType())
		parent := je1.getParentAsJSONObject()
		require.Equal(t, jsonObj, parent)
		log.Logger.Infof("%+v", parent.String())

		je2 := parent.getAsJSONElement("key2")
		log.Logger.Infof("%+v", je2.String())
		require.Equal(t, 3.14, je2.getValue())

		log.Logger.Infof("%v", jsonObj.GetAsJSON())
	})

	t.Run("Can put JSONObject to JSONObject", func(t *testing.T) {
		opID1 := model.NewOperationID()
		jsonObj := newJSONObject(nil, model.OldestTimestamp)

		strt := struct {
			K1_1 string
			K1_2 int
		}{
			K1_1: "hello",
			K1_2: 1234,
		}

		jsonObj.put("K1", strt, opID1.Next().GetTimestamp())
		log.Logger.Infof("%v", jsonObj)
		k1JSONObj := jsonObj.getAsJSONObject("K1")
		log.Logger.Infof("%v", k1JSONObj)

		// add struct ptr
		k1JSONObj.put("K1_3", &strt, opID1.Next().GetTimestamp())

		parent := k1JSONObj.getParentAsJSONObject()
		log.Logger.Infof("%v", parent)
		require.Equal(t, jsonObj, parent)
		log.Logger.Infof("%v", jsonObj.GetAsJSON())
		d, err := json.Marshal(jsonObj.GetAsJSON())
		require.NoError(t, err)
		log.Logger.Infof("%s", string(d))
		require.Equal(t, string(d), `{"K1":{"K1_1":"hello","K1_2":1234,"K1_3":{"K1_1":"hello","K1_2":1234}}}`)
	})

	t.Run("Can put JSONArray to JSONObject", func(t *testing.T) {
		opID1 := model.NewOperationID()
		jsonObj := newJSONObject(nil, model.OldestTimestamp)

		array := []interface{}{123, 3.14, "hello"}
		jsonObj.put("K1", array, opID1.Next().GetTimestamp())
		log.Logger.Infof("%v", jsonObj.String())

	})

	// t.Run("Can test JSON operations", func(t *testing.T) {
	// 	snap := newJSONSnapshot()
	// 	hello := struct {
	// 		X string
	// 		Y int32
	// 	}{
	// 		X: "hi",
	// 		Y: 999,
	// 	}
	// 	world := struct {
	// 		A int32
	// 		B string
	// 		C struct {
	// 			X float32
	// 		}
	// 		D []interface{}
	// 	}{
	// 		A: 10,
	// 		B: "string",
	// 		C: struct{ X float32 }{X: 3.141592},
	// 		D: []interface{}{"world", 1, 3.141592, hello},
	// 	}
	// 	op1 := model.NewOperationID()
	// 	snap.putLocal("key1", world, op1.Next().GetTimestamp())
	//
	// 	// world := []interface{}{"world", 1, 3.141592, world}
	// 	// snap.putLocal("key2", world, op1.Next().GetTimestamp())
	// 	// snap.putLocal("key2", 123, op1.Next().GetTimestamp())
	// 	// snap.putLocal("key3", []string{"a", "b", "c"}, op1.Next().GetTimestamp())
	// })
	//
	// t.Run("Can convert type to JSON related object ", func(t *testing.T) {
	// 	snap := newJSONSnapshot()
	// 	require.Equal(t, int64(1), snap.convertJSONType(1))
	// 	require.Equal(t, 3.141592, snap.convertJSONType(3.141592))
	// 	require.Equal(t, "hello", snap.convertJSONType("hello"))
	// 	var strPtr = "world"
	// 	require.Equal(t, "world", snap.convertJSONType(&strPtr))
	// 	var intVal = 12345
	// 	require.Equal(t, int64(12345), snap.convertJSONType(&intVal))
	//
	// })
}
