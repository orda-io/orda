package orda

import (
	"github.com/orda-io/orda/client/pkg/errors"
	"github.com/orda-io/orda/client/pkg/log"
	"github.com/orda-io/orda/client/pkg/model"
	testonly2 "github.com/orda-io/orda/client/pkg/testonly"
	"github.com/orda-io/orda/client/pkg/utils"
	"github.com/wI2L/jsondiff"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	arr  = []interface{}{"a", 2}
	str1 = struct {
		E1 string
		E2 int
		A3 []interface{}
	}{
		E1: "hello",
		E2: 1234,
		A3: arr,
	}
)

func TestDocument(t *testing.T) {
	t.Run("Can test operations of JSONObject Document", func(t *testing.T) {
		tw := testonly2.NewTestWire(true)
		root, _ := newDocument(testonly2.NewBase(t.Name(), model.TypeOfDatatype_DOCUMENT), tw, nil)

		old1, oErr := root.PutToObject("K1", "V1")
		require.NoError(t, oErr)
		require.Nil(t, old1)
		require.False(t, root.IsGarbage())

		old2, oErr := root.PutToObject("K1", "V2")
		require.NoError(t, oErr)
		require.NotNil(t, old2)
		require.True(t, old2.IsGarbage())
		require.Equal(t, TypeJSONElement, old2.GetTypeOfJSON())
		require.Equal(t, "V1", old2.ToJSON())

		// obtain updated key
		k1, oErr := root.GetFromObject("K1")
		require.NoError(t, oErr)
		require.False(t, k1.IsGarbage())
		require.Equal(t, "V2", k1.ToJSON())

		// delete updated key
		old3, oErr := root.DeleteInObject("K1")
		require.NoError(t, oErr)
		require.NotNil(t, old3)
		require.True(t, old3.IsGarbage())
		require.Equal(t, TypeJSONElement, old3.GetTypeOfJSON())
		require.Equal(t, "V2", old3.ToJSON())

		// obtain deleted key
		k1, oErr = root.GetFromObject("K1")
		require.NoError(t, oErr)
		require.Nil(t, k1)

		// put struct into Document
		old4, oErr := root.PutToObject("K2", str1)
		require.NoError(t, oErr)
		require.Nil(t, old4)

		k2, oErr := root.GetFromObject("K2")
		require.NoError(t, oErr)
		require.Equal(t, TypeJSONObject, k2.GetTypeOfJSON())

		oldK2, oErr := k2.PutToObject("E1", "V3")
		require.NoError(t, oErr)
		require.True(t, oldK2.IsGarbage())
		e1, oErr := k2.GetFromObject("E1")
		require.NoError(t, oErr)
		require.Equal(t, "V3", e1.ToJSON())

		log.Logger.Infof("%v", testonly2.Marshal(t, root.ToJSON()))

		oldK3, oErr := root.PutToObject("K2", "V4")
		require.NoError(t, oErr)
		require.True(t, oldK3.IsGarbage())
		require.True(t, k2.IsGarbage())

		// put on deleted Document
		oldE4, oErr := oldK3.PutToObject("E4", "V4")
		require.Error(t, oErr)
		require.Equal(t, errors.DatatypeNoOp, oErr.GetCode())
		require.Nil(t, oldE4)

		// delete on deleted Document
		oldE4a, oErr := k2.DeleteInObject("E4")
		require.Error(t, oErr)
		require.Equal(t, errors.DatatypeNoOp, oErr.GetCode())
		require.Nil(t, oldE4a)

		log.Logger.Infof("%v", testonly2.Marshal(t, root.ToJSON()))

		require.True(t, root.GetRootDocument().Equal(k2.GetRootDocument()))

		opID1 := root.(*document).GetOpID().Clone()
		not, oErr := root.DeleteInObject("NOT_EXISTING")
		require.Error(t, oErr)
		require.Equal(t, errors.DatatypeNoOp, oErr.GetCode())
		require.Nil(t, not)
		opID2 := root.(*document).GetOpID().Clone()
		require.Equal(t, 0, opID1.Compare(opID2))
	})

	t.Run("Can test operations of JSONArray Document", func(t *testing.T) {
		tw := testonly2.NewTestWire(true)
		root, _ := newDocument(testonly2.NewBase(t.Name(), model.TypeOfDatatype_DOCUMENT), tw, nil)

		old1, oErr := root.PutToObject("K1", arr)
		require.NoError(t, oErr)
		require.Nil(t, old1)

		array, oErr := root.GetFromObject("K1")
		require.NoError(t, oErr)
		require.NotNil(t, array)

		// test InsertToArray
		arr1, oErr := array.InsertToArray(0, "x", "y")
		require.NoError(t, oErr)
		require.True(t, array.Equal(arr1))

		opID1 := root.(*document).GetOpID().Clone()
		arr2, oErr := array.InsertToArray(100, "x", "y")
		require.Error(t, oErr)
		require.True(t, array.Equal(arr2))
		opID2 := root.(*document).GetOpID().Clone()
		require.Equal(t, 0, opID1.Compare(opID2))

		// test UpdateManyInArray
		existing, oErr := array.UpdateManyInArray(1, "Y", "A")
		require.NoError(t, oErr)
		require.True(t, existing[0].IsGarbage())
		require.True(t, existing[1].IsGarbage())
		require.Equal(t, "y", existing[0].ToJSON())
		require.Equal(t, "a", existing[1].ToJSON())
		log.Logger.Infof("%v", testonly2.Marshal(t, root.ToJSON()))

		// test DeleteInArray
		deleted, oErr := array.DeleteManyInArray(0, 2)
		require.NoError(t, oErr)
		require.Equal(t, 2, len(deleted))
		require.Equal(t, "x", deleted[0].ToJSON())
		require.Equal(t, "Y", deleted[1].ToJSON())
		require.True(t, deleted[0].IsGarbage())
		require.True(t, deleted[1].IsGarbage())

		arr3, oErr := array.InsertToArray(2, arr)
		require.NoError(t, oErr)
		require.True(t, array.Equal(arr3))

		insArr, oErr := array.GetFromArray(2)
		require.NoError(t, oErr)
		require.True(t, array.Equal(insArr.GetParentDocument()))
		require.False(t, insArr.IsGarbage())

		oldArr, oErr := array.UpdateManyInArray(2, "X")
		require.NoError(t, oErr)
		require.True(t, oldArr[0].IsGarbage())
		require.True(t, insArr.Equal(oldArr[0]))

		sameArr, oErr := insArr.InsertToArray(0, "X")
		require.Error(t, oErr)
		require.Equal(t, oErr.GetCode(), errors.DatatypeNoOp)
		require.True(t, insArr.Equal(sameArr))

		log.Logger.Infof("%v", testonly2.Marshal(t, root.ToJSON()))
	})

	t.Run("Can transaction for Document", func(t *testing.T) {
		tw := testonly2.NewTestWire(true)

		outDoc, _ := newDocument(testonly2.NewBase(t.Name(), model.TypeOfDatatype_DOCUMENT), tw, nil)

		err := outDoc.Transaction("transaction1", func(doc DocumentInTx) error {
			_, err := doc.PutToObject("K1", "V1")
			require.NoError(t, err)
			_, err = doc.PutToObject("K2", str1)
			require.NoError(t, err)
			_, err = doc.GetFromObject("K2")
			require.NoError(t, err)
			// log.Logger.Infof("%v", testonly.Marshal(t, doc.GetAsJSON()))
			// _, _ = counter.IncreaseBy(2)
			// require.Equal(t, int32(2), outDoc.Get())
			// _, _ = counter.IncreaseBy(4)
			// require.Equal(t, int32(6), counter.Get())
			return nil
		})
		require.NoError(t, err)

		// require.Equal(t, int32(6), outDoc.Get())
		//
		// require.Error(t, outDoc.DoTransaction("transaction2", func(intCounter CounterInTx) error {
		// 	_, _ = intCounter.IncreaseBy(3)
		// 	require.Equal(t, int32(9), intCounter.Get())
		// 	_, _ = intCounter.IncreaseBy(5)
		// 	require.Equal(t, int32(14), intCounter.Get())
		// 	return fmt.Errorf("err")
		// }))
		// require.Equal(t, int32(6), outDoc.Get())

	})

	t.Run("Can execute PatchByJSON", func(t *testing.T) {
		tw := testonly2.NewTestWire(true)
		original, _ := newDocument(testonly2.NewBase(t.Name(), model.TypeOfDatatype_DOCUMENT), tw, nil)
		byPatch, _ := newDocument(testonly2.NewBase(t.Name(), model.TypeOfDatatype_DOCUMENT), tw, nil)
		_, err := original.PutToObject("k1", str1)
		require.NoError(t, err)
		_, err = original.PutToObject("k2", arr1)
		require.NoError(t, err)

		utils.PrintMarshalDoc(log.Logger, original.ToJSON())
		_, err = byPatch.PatchByJSON(utils.ToStringMarshalDoc(original.ToJSON()))
		require.NoError(t, err)
		utils.PrintMarshalDoc(log.Logger, byPatch.ToJSON())
		utils.PrintMarshalDoc(log.Logger, original.ToJSON())

		require.Equal(t, utils.ToStringMarshalDoc(original), utils.ToStringMarshalDoc(byPatch))
	})

	t.Run("Can execute Patch", func(t *testing.T) {

		obj1 := struct {
			K1 string
			K2 int
			K3 bool
		}{
			K1: "hello",
			K2: 1234,
			K3: true,
		}
		arr1 := []interface{}{"world", 1234, 3.141592, false}
		tw := testonly2.NewTestWire(true)
		root, _ := newDocument(testonly2.NewBase(t.Name(), model.TypeOfDatatype_DOCUMENT), tw, nil)
		_, err := root.PutToObject("objKey", obj1)
		require.NoError(t, err)
		_, err = root.PutToObject("arrKey", arr1)
		require.NoError(t, err)

		p1 := jsondiff.Operation{
			Type:  "add",
			Path:  "/objKey",
			Value: arr1,
		}
		p2 := jsondiff.Operation{
			Type: "remove",
			Path: "/objKey/K1",
		}
		p3 := jsondiff.Operation{
			Type:  "replace",
			Path:  "/objKey/K2",
			Value: 5678,
		}

		original := utils.ToStringMarshalDoc(root.ToJSON())

		// should fail patch transaction
		require.Error(t, root.Patch(p1, p2, p3))
		require.Equal(t, original, utils.ToStringMarshalDoc(root.ToJSON()))

		p1.Path = "/objKey/newKey"
		require.NoError(t, root.Patch(p1, p2, p3))

		utils.PrintMarshalDoc(log.Logger, root.ToJSON())

		getObj1, err := root.GetByPath("/")
		require.NoError(t, err)
		require.True(t, getObj1.Equal(root))

		newKey, err := root.GetByPath("/objKey/newKey")
		require.NoError(t, err)
		require.Equal(t, utils.ToStringMarshalDoc(newKey.ToJSON()), utils.ToStringMarshalDoc(arr1))

		removeKey, err := root.GetByPath("objKey/K1")
		require.Error(t, err)
		require.Nil(t, removeKey)

		replaceKey, err := root.GetByPath("objKey/K2")
		require.NoError(t, err)
		require.Equal(t, float64(5678), replaceKey.GetValue())

		p4 := jsondiff.Operation{
			Path:  "/arrKey/0",
			Type:  "add",
			Value: obj1,
		}

		p5 := jsondiff.Operation{
			Path: "/arrKey/1",
			Type: "remove",
		}

		p6 := jsondiff.Operation{
			Path:  "/arrKey/1",
			Type:  "replace",
			Value: 5678,
		}
		require.NoError(t, root.Patch(p4, p5, p6))
		utils.PrintMarshalDoc(log.Logger, root.ToJSON())

		arrKey0, err := root.GetByPath("/arrKey/0")
		require.NoError(t, err)
		require.Equal(t, utils.ToStringMarshalDoc(arrKey0.ToJSON()), utils.ToStringMarshalDoc(obj1))

		arrKey, err := root.GetByPath("/arrKey")
		require.NoError(t, err)
		require.Equal(t, len(arrKey.ToJSON().([]interface{})), 4)

		arrKey1, err := root.GetByPath("/arrKey/1")
		require.NoError(t, err)
		require.Equal(t, float64(5678), arrKey1.GetValue())
	})

	t.Run("Can patch nested json", func(t *testing.T) {
		json := "{\"k1\":[ {\"k1-0-1\":\"v1\",\"k1-0-2\":\"v2\"}, {\"k1-1-1\":\"v3\",\"k1-1-2\":\"v4\"} ],\"k2\":[ {\"k2-0-1\":\"v5\",\"k2-0-2\":\"v6\"}, {\"k2-1-1\":\"v7\",\"k2-1-2\":\"v8\"} ] }"
		tw := testonly2.NewTestWire(true)
		root, _ := newDocument(testonly2.NewBase(t.Name(), model.TypeOfDatatype_DOCUMENT), tw, nil)
		patches, err := root.PatchByJSON(json)
		require.NoError(t, err)
		require.Equal(t, 2, len(patches))

		elem, err := root.GetByPath("/k1/0/k1-0-1")
		require.NoError(t, err)
		require.Equal(t, "v1", elem.GetValue())

		obj, err := root.GetByPath("/k1/1")
		require.NoError(t, err)
		require.Equal(t, TypeJSONObject, obj.GetTypeOfJSON())

		elem2, err := obj.GetFromObject("k1-1-1")
		require.NoError(t, err)
		require.Equal(t, elem2.GetValue(), "v3")
	})
}
