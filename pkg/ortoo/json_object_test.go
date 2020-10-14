package ortoo

import (
	"encoding/json"
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/log"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/pkg/testonly"
	"github.com/stretchr/testify/require"
	"testing"
)

func initJSONObjectAndTestPut(
	t *testing.T,
	root *jsonObject,
	opID *model.OperationID,
) []*model.Timestamp {
	// { "K1": "hello" }
	TS1 := opID.Next().GetTimestamp()
	removed1 := root.putCommon("K1", "hello", TS1.Clone())
	require.Nil(t, removed1)
	put1 := root.getAsJSONType("K1")
	require.Equal(t, root, put1.getParent())
	require.Equal(t, TS1, put1.getCreateTime())

	// { "K1": "hello", "K2": { "K1": "hello", "K2": 1234 } }
	TS2 := opID.Next().GetTimestamp()
	removed2 := root.putCommon("K2", strt1, TS2.Clone())
	require.Nil(t, removed2)
	put2 := root.getAsJSONType("K2")
	require.Equal(t, root, put2.getParent())
	require.Equal(t, TS2, put2.getCreateTime())

	// { "K1": "hello", "K2": { "K1": "hello", "K2": 1234 }, "K3": ["world", 1234, 3.14] }
	TS3 := opID.Next().GetTimestamp()
	removed3 := root.putCommon("K3", arr1, TS3.Clone())
	require.Nil(t, removed3)
	put3 := root.getAsJSONType("K3")
	require.Equal(t, root, put3.getParent())
	require.Equal(t, TS3, put3.getCreateTime())

	log.Logger.Infof("%v", testonly.Marshal(t, root.GetAsJSONCompatible()))
	return []*model.Timestamp{TS1, TS2, TS3}
}

func jsonObjectMarshalTest(t *testing.T, original *jsonObject) {
	clone := newJSONObject(original.base, nil, model.OldestTimestamp())
	m, err := json.Marshal(original) // ==> jsonObject.MarshalJSON
	require.NoError(t, err)

	err = json.Unmarshal(m, clone)
	require.NoError(t, err)
	require.True(t, original.Equal(clone))
}

func TestJSONObject(t *testing.T) {
	t.Run("Can put JSONElements to JSONObject and obtain the parent of children", func(t *testing.T) {
		opID := model.NewOperationID()
		base := testonly.NewBase(t.Name(), model.TypeOfDatatype_DOCUMENT)
		root := newJSONObject(base, nil, model.OldestTimestamp())

		// {"K1":1234,"K2":{"K1":"hello","K2":1234},"K3":["world",1234,3.14]}
		root.putCommon("K1", 1234, opID.Next().GetTimestamp())
		root.putCommon("K2", strt1, opID.Next().GetTimestamp())
		root.putCommon("K3", arr1, opID.Next().GetTimestamp())
		log.Logger.Infof("%v", testonly.Marshal(t, root.GetAsJSONCompatible()))

		// get the child K1
		child1 := root.getAsJSONType("K1")
		require.Equal(t, float64(1234), child1.getValue())
		require.Equal(t, TypeJSONObject, child1.getParent().getType())
		require.Equal(t, root, child1.getParent())

		// get parent of K2
		child2 := root.getAsJSONType("K2").(*jsonObject)
		require.Equal(t, TypeJSONObject, child2.getParent().getType())
		require.Equal(t, root, child2.getParent())
		require.Equal(t, child2, child2.getAsJSONType("K1").getParent())

		// get parent of K3
		child3 := root.getAsJSONType("K3").(*jsonArray)
		require.Equal(t, TypeJSONObject, child3.getParent().getType())
		require.Equal(t, root, child3.getParent())
		grandChild3 := child3.findTimedType(0)
		require.Equal(t, child3, grandChild3.(*jsonElement).getParent())

		// marshal and unmarshal snapshot
		jsonObjectMarshalTest(t, root)
	})

	t.Run("Can put JSONArray to JSONObject", func(t *testing.T) {
		opID := model.NewOperationID()
		base := testonly.NewBase(t.Name(), model.TypeOfDatatype_DOCUMENT)
		root := newJSONObject(base, nil, model.OldestTimestamp())

		// put array for K1
		array := []interface{}{1234, 3.14, "hello"}
		old, oErr := root.PutCommonInObject(root.getCreateTime(), "K1", array, opID.Next().GetTimestamp())
		require.NoError(t, oErr)
		require.Nil(t, old)
		log.Logger.Infof("%v", testonly.Marshal(t, root.GetAsJSONCompatible()))

		// get jsonArray from jsonObject
		arr := root.getChildAsJSONArray("K1")
		require.Equal(t, 3, arr.Size())
		require.Equal(t, root, arr.getParent())

		// get jsonElement from jsonArray
		val1 := arr.findTimedType(0)
		require.Equal(t, float64(1234), val1.getValue())
		require.Equal(t, arr, val1.(*jsonElement).getParent())

		// marshal and unmarshal snapshot
		jsonObjectMarshalTest(t, root)
	})

	t.Run("Can update values in JSONObject", func(t *testing.T) {
		opID := model.NewOperationID()
		base := testonly.NewBase(t.Name(), model.TypeOfDatatype_DOCUMENT)
		root := newJSONObject(base, nil, model.OldestTimestamp())

		initJSONObjectAndTestPut(t, root, opID)

		// Replace an existing JSONElement with a new JSONElement.
		// The existing JSONElement will be deleted
		old1, err := root.PutCommonInObject(root.getCreateTime(), "K1", "update1", opID.Next().GetTimestamp())
		require.NoError(t, err)
		require.True(t, old1.isTomb())
		require.Equal(t, "hello", old1.getValue())
		log.Logger.Infof("%v", testonly.Marshal(t, root.GetAsJSONCompatible()))

		// Replace an existing JSONObject with a new JSONElement
		old2, err := root.PutCommonInObject(root.getCreateTime(), "K2", "update2", opID.Next().GetTimestamp())
		require.NoError(t, err)
		require.Equal(t, TypeJSONObject, old2.getType())
		require.True(t, old2.isTomb())
		log.Logger.Infof("%v", testonly.Marshal(t, root.GetAsJSONCompatible()))

		// Update an already deleted JSONObject
		// In this case, the operation is applied effectively, but it is not shown in the root document.
		old3, err := root.PutCommonInObject(old2.getCreateTime(), "K1", "update3", opID.Next().GetTimestamp())
		require.NoError(t, err)
		require.Equal(t, "hello", old3.getValue())
		require.True(t, old3.isTomb())
		require.Equal(t, "update3", old2.(*jsonObject).getAsJSONType("K1").getValue())
		log.Logger.Infof("%v", testonly.Marshal(t, old2.(*jsonObject).GetAsJSONCompatible()))

		// Replace an existing JSONArray with a new JSONObject
		old4, err := root.PutCommonInObject(root.getCreateTime(), "K3", strt1, opID.Next().GetTimestamp())
		require.NoError(t, err)
		require.Equal(t, TypeJSONArray, old4.getType())
		require.True(t, old4.isTomb())
		log.Logger.Infof("%v", testonly.Marshal(t, root.GetAsJSONCompatible()))

		// marshal and unmarshal snapshot
		jsonObjectMarshalTest(t, root)
	})

	t.Run("Can delete something locally in JSONObject", func(t *testing.T) {
		opID := model.NewOperationID()
		base := testonly.NewBase(t.Name(), model.TypeOfDatatype_DOCUMENT)
		root := newJSONObject(base, nil, model.OldestTimestamp())

		initJSONObjectAndTestPut(t, root, opID)

		// delete not existing
		old0, err := root.DeleteCommonInObject(root.getCreateTime(), "NOT_EXISTING", opID.Next().GetTimestamp(), true)
		require.Error(t, err)
		require.Equal(t, errors.DatatypeNoOp.ToErrorCode(), err.GetCode())
		require.Nil(t, old0)

		// delete a jsonElement
		old1, err := root.DeleteCommonInObject(root.getCreateTime(), "K1", opID.Next().GetTimestamp(), true)
		require.NoError(t, err)
		require.Equal(t, "hello", old1.getValue())
		require.True(t, old1.isTomb()) // should be tombstone.
		log.Logger.Infof("%v", testonly.Marshal(t, root.GetAsJSONCompatible()))

		// delete again: it should be ignored.
		old2, err := root.DeleteCommonInObject(root.getCreateTime(), "K1", opID.Next().GetTimestamp(), true)
		require.Error(t, err)
		require.Equal(t, errors.DatatypeNoOp.ToErrorCode(), err.GetCode())
		require.Nil(t, old2)
		log.Logger.Infof("%v", testonly.Marshal(t, root.GetAsJSONCompatible()))

		// delete a JSONObject
		old3, err := root.DeleteCommonInObject(root.getCreateTime(), "K2", opID.Next().GetTimestamp(), true)
		require.NoError(t, err)
		require.True(t, old3.isTomb())
		log.Logger.Infof("%v", testonly.Marshal(t, old3.(*jsonObject).GetAsJSONCompatible()))
		log.Logger.Infof("%v", testonly.Marshal(t, root.GetAsJSONCompatible()))

		// delete a JSONArray
		old4, err := root.DeleteCommonInObject(root.getCreateTime(), "K3", opID.Next().GetTimestamp(), true)
		require.NoError(t, err)
		require.True(t, old4.isTomb())
		log.Logger.Infof("%v", testonly.Marshal(t, old4.(*jsonArray).GetAsJSONCompatible()))
		log.Logger.Infof("%v", testonly.Marshal(t, root.GetAsJSONCompatible()))

		// marshal and unmarshal snapshot
		jsonObjectMarshalTest(t, root)
	})

	t.Run("Can delete something remotely in JSONObject", func(t *testing.T) {
		opID := model.NewOperationID()
		base := testonly.NewBase(t.Name(), model.TypeOfDatatype_DOCUMENT)
		root := newJSONObject(base, nil, model.OldestTimestamp())

		ts := initJSONObjectAndTestPut(t, root, opID)

		// delete NOT_EXISTING remotely
		old0, err := root.DeleteCommonInObject(root.getCreateTime(), "NOT_EXISTING", opID.Next().GetTimestamp(), false)
		require.Nil(t, old0)
		require.Error(t, err)
		require.Equal(t, errors.DatatypeNoTarget.ToErrorCode(), err.GetCode())

		// delete a JSONElement remotely
		old1, err := root.DeleteCommonInObject(root.getCreateTime(), "K1", opID.Next().GetTimestamp(), false)
		require.NoError(t, err)
		require.True(t, old1.isTomb())
		require.Equal(t, opID.GetTimestamp(), old1.getDeleteTime())
		require.Equal(t, ts[0], old1.getCreateTime())
		log.Logger.Infof("%v", testonly.Marshal(t, root.GetAsJSONCompatible()))

		// if deleting again, newer timestamp is replaced with the previous one.
		old2, err := root.DeleteCommonInObject(root.getCreateTime(), "K1", opID.Next().GetTimestamp(), false)
		require.NoError(t, err)
		require.True(t, old2.isTomb())
		require.Equal(t, ts[0], old2.getCreateTime())
		require.Equal(t, opID.GetTimestamp(), old2.getDeleteTime())
		require.Equal(t, opID.GetTimestamp(), root.getChildAsJSONElement("K1").getDeleteTime())
		log.Logger.Infof("%v", testonly.Marshal(t, root.GetAsJSONCompatible()))

		// delete an JSONObject
		old3, err := root.DeleteCommonInObject(root.getCreateTime(), "K2", opID.Next().GetTimestamp(), false)
		require.NoError(t, err)
		require.True(t, old3.isTomb())
		child3 := old3.(*jsonObject).getAsJSONType("K1")
		require.False(t, child3.isTomb())
		log.Logger.Infof("%v", testonly.Marshal(t, old3.(*jsonObject).GetAsJSONCompatible()))
		log.Logger.Infof("%v", testonly.Marshal(t, root.GetAsJSONCompatible()))

		// delete an JSONArray
		old4, err := root.DeleteCommonInObject(root.getCreateTime(), "K3", opID.Next().GetTimestamp(), false)
		require.NoError(t, err)
		require.True(t, old4.isTomb())
		child4 := old4.(*jsonArray).getJSONType(0)
		require.False(t, child4.isTomb())
		require.Equal(t, opID.GetTimestamp(), old4.getDeleteTime())
		log.Logger.Infof("%v", testonly.Marshal(t, root.GetAsJSONCompatible()))

		// delete the deleted JSONArray with older timestamp
		// It is ignored.
		old5, err := root.DeleteCommonInObject(root.getCreateTime(), "K3", ts[0], false)
		require.NoError(t, err)
		require.Nil(t, old5)
		require.NotEqual(t, ts[0], old4.getDeleteTime())

		// marshal and unmarshal snapshot
		jsonObjectMarshalTest(t, root)
	})
}
