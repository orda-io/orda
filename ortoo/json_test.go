package ortoo

import (
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/ortoo/model"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestJSONSnapshot(t *testing.T) {

	t.Run("...", func(t *testing.T) {
		opID1 := model.NewOperationID()
		jsonObject := newJSONObject()
		jsonObject.put("key1", 1234, opID1.Next().GetTimestamp())
		jsonObject.put("key2", 3.141592, opID1.Next().GetTimestamp())
		je := jsonObject.getAsJSONElement("key1")
		log.Logger.Infof("%+v", je)
		parent := je.getParentAsJSONObject()
		require.Equal(t, typeJSONObject, je.getParent().getType())
		log.Logger.Infof("%+v", parent)

	})

	t.Run("Can test JSON operations", func(t *testing.T) {
		snap := newJSONSnapshot()
		hello := struct {
			X string
			Y int32
		}{
			X: "hi",
			Y: 999,
		}
		world := struct {
			A int32
			B string
			C struct {
				X float32
			}
			D []interface{}
		}{
			A: 10,
			B: "string",
			C: struct{ X float32 }{X: 3.141592},
			D: []interface{}{"world", 1, 3.141592, hello},
		}
		op1 := model.NewOperationID()
		snap.putLocal("key1", world, op1.Next().GetTimestamp())

		// world := []interface{}{"world", 1, 3.141592, world}
		// snap.putLocal("key2", world, op1.Next().GetTimestamp())
		// snap.putLocal("key2", 123, op1.Next().GetTimestamp())
		// snap.putLocal("key3", []string{"a", "b", "c"}, op1.Next().GetTimestamp())
	})

	t.Run("Can convert type to JSON related object ", func(t *testing.T) {
		snap := newJSONSnapshot()
		require.Equal(t, int64(1), snap.convertJSONType(1))
		require.Equal(t, 3.141592, snap.convertJSONType(3.141592))
		require.Equal(t, "hello", snap.convertJSONType("hello"))
		var strPtr = "world"
		require.Equal(t, "world", snap.convertJSONType(&strPtr))
		var intVal = 12345
		require.Equal(t, int64(12345), snap.convertJSONType(&intVal))

	})
}
