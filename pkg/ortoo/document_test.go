package ortoo

import (
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/log"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/pkg/testonly"
	"github.com/stretchr/testify/require"
	"testing"
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
		tw := testonly.NewTestWire(true)
		root, _ := newDocument(testonly.NewBase(t.Name(), model.TypeOfDatatype_DOCUMENT), tw, nil)

		old1, oErr := root.PutToObject("K1", "V1")
		require.NoError(t, oErr)
		require.Nil(t, old1)
		require.False(t, root.IsGarbage())

		old2, oErr := root.PutToObject("K1", "V2")
		require.NoError(t, oErr)
		require.NotNil(t, old2)
		require.True(t, old2.IsGarbage())
		require.Equal(t, TypeJSONElement, old2.GetJSONType())
		require.Equal(t, "V1", old2.GetAsJSON())

		// obtain updated key
		k1, oErr := root.GetFromObject("K1")
		require.NoError(t, oErr)
		require.False(t, k1.IsGarbage())
		require.Equal(t, "V2", k1.GetAsJSON())

		// delete updated key
		old3, oErr := root.DeleteInObject("K1")
		require.NoError(t, oErr)
		require.NotNil(t, old3)
		require.True(t, old3.IsGarbage())
		require.Equal(t, TypeJSONElement, old3.GetJSONType())
		require.Equal(t, "V2", old3.GetAsJSON())

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
		require.Equal(t, TypeJSONObject, k2.GetJSONType())

		oldK2, oErr := k2.PutToObject("E1", "V3")
		require.NoError(t, oErr)
		require.True(t, oldK2.IsGarbage())
		e1, oErr := k2.GetFromObject("E1")
		require.NoError(t, oErr)
		require.Equal(t, "V3", e1.GetAsJSON())

		log.Logger.Infof("%v", testonly.Marshal(t, root.GetAsJSON()))

		oldK3, oErr := root.PutToObject("K2", "V4")
		require.NoError(t, oErr)
		require.True(t, oldK3.IsGarbage())
		require.True(t, k2.IsGarbage())

		// put on deleted Document
		oldE4, oErr := oldK3.PutToObject("E4", "V4")
		require.Error(t, oErr)
		require.Equal(t, errors.DatatypeNoOp.ToErrorCode(), oErr.GetCode())
		require.Nil(t, oldE4)

		// delete on deleted Document
		oldE4a, oErr := k2.DeleteInObject("E4")
		require.Error(t, oErr)
		require.Equal(t, errors.DatatypeNoOp.ToErrorCode(), oErr.GetCode())
		require.Nil(t, oldE4a)

		log.Logger.Infof("%v", testonly.Marshal(t, root.GetAsJSON()))

		require.True(t, root.GetRootDocument().Equal(k2.GetRootDocument()))

		opID1 := root.(*document).GetBase().GetOpID().Clone()
		not, oErr := root.DeleteInObject("NOT_EXISTING")
		require.Error(t, oErr)
		require.Equal(t, errors.DatatypeNoOp.ToErrorCode(), oErr.GetCode())
		require.Nil(t, not)
		opID2 := root.(*document).GetBase().GetOpID().Clone()
		require.Equal(t, 0, opID1.Compare(opID2))
	})

	t.Run("Can test operations of JSONArray Document", func(t *testing.T) {
		tw := testonly.NewTestWire(true)
		root, _ := newDocument(testonly.NewBase(t.Name(), model.TypeOfDatatype_DOCUMENT), tw, nil)

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

		opID1 := root.(*document).GetBase().GetOpID().Clone()
		arr2, oErr := array.InsertToArray(100, "x", "y")
		require.Error(t, oErr)
		require.True(t, array.Equal(arr2))
		opID2 := root.(*document).GetBase().GetOpID().Clone()
		require.Equal(t, 0, opID1.Compare(opID2))

		// test UpdateManyInArray
		existing, oErr := array.UpdateManyInArray(1, "Y", "A")
		require.NoError(t, oErr)
		require.True(t, existing[0].IsGarbage())
		require.True(t, existing[1].IsGarbage())
		require.Equal(t, "y", existing[0].GetAsJSON())
		require.Equal(t, "a", existing[1].GetAsJSON())
		log.Logger.Infof("%v", testonly.Marshal(t, root.GetAsJSON()))

		// test DeleteInArray
		deleted, oErr := array.DeleteManyInArray(0, 2)
		require.NoError(t, oErr)
		require.Equal(t, 2, len(deleted))
		require.Equal(t, "x", deleted[0].GetAsJSON())
		require.Equal(t, "Y", deleted[1].GetAsJSON())
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
		require.Equal(t, oErr.GetCode(), errors.DatatypeNoOp.ToErrorCode())
		require.True(t, insArr.Equal(sameArr))

		log.Logger.Infof("%v", testonly.Marshal(t, root.GetAsJSON()))
	})

	t.Run("Can transaction for Document", func(t *testing.T) {
		tw := testonly.NewTestWire(true)

		outDoc, _ := newDocument(testonly.NewBase(t.Name(), model.TypeOfDatatype_DOCUMENT), tw, nil)

		err := outDoc.DoTransaction("transaction1", func(doc DocumentInTxn) error {
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
		// require.Error(t, outDoc.DoTransaction("transaction2", func(intCounter CounterInTxn) error {
		// 	_, _ = intCounter.IncreaseBy(3)
		// 	require.Equal(t, int32(9), intCounter.Get())
		// 	_, _ = intCounter.IncreaseBy(5)
		// 	require.Equal(t, int32(14), intCounter.Get())
		// 	return fmt.Errorf("err")
		// }))
		// require.Equal(t, int32(6), outDoc.Get())

	})
}
