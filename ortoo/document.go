package ortoo

import (
	"fmt"
	"github.com/knowhunger/ortoo/ortoo/errors"
	"github.com/knowhunger/ortoo/ortoo/iface"
	"github.com/knowhunger/ortoo/ortoo/internal/datatypes"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/ortoo/model"
	"github.com/knowhunger/ortoo/ortoo/operations"
	"github.com/knowhunger/ortoo/ortoo/types"
	"reflect"
)

type Document interface {
	Add(key string, value interface{})
}

func newDocument(key string, cuid types.CUID, wire iface.Wire, handlers *Handlers) Document {
	doc := &document{
		datatype: &datatype{
			ManageableDatatype: &datatypes.ManageableDatatype{},
			handlers:           handlers,
		},
		snapshot: nil,
	}
	doc.Initialize(key, model.TypeOfDatatype_JSON, cuid, wire, doc.snapshot, doc)
	return doc
}

type document struct {
	*datatype
	snapshot *jsonObject
}

func (its *document) Add(key string, value interface{}) {
	op := operations.NewAddOperation(key, value)
	ret, err := its.ExecuteOperationWithTransaction(its.TransactionCtx, op, true)
	if err != nil {
		// return nil,
	}
	// return ret, nil
}

func (its *document) ExecuteLocal(op interface{}) (interface{}, error) {
	switch cast := op.(type) {
	case *operations.AddOperation:
	case *operations.CutOperation:
	case *operations.SetOperation:
	}
	return nil, errors.NewDatatypeError(errors.ErrDatatypeIllegalOperation, op)
}

func (its *document) ExecuteRemote(op interface{}) (interface{}, error) {
	switch cast := op.(type) {
	case *operations.SnapshotOperation:
	case *operations.AddOperation:
	case *operations.CutOperation:
	case *operations.SetOperation:
	}
	return nil, errors.NewDatatypeError(errors.ErrDatatypeIllegalOperation, op)
}

func (its *document) GetAsJSON() interface{} {
	panic("implement me")
}

func (its *document) SetSnapshot(snapshot iface.Snapshot) {
	panic("implement me")
}

func (its *document) GetSnapshot() iface.Snapshot {
	panic("implement me")
}

func (its *document) GetMetaAndSnapshot() ([]byte, iface.Snapshot, error) {
	panic("implement me")
}

func (its *document) SetMetaAndSnapshot(meta []byte, snapshot string) error {
	panic("implement me")
}

// ////////////////////////////////////////////////////////////////
//  jsonSnapshot
// ////////////////////////////////////////////////////////////////

type TypeOfJSON int

const (
	TypeJSONPrimitive TypeOfJSON = iota
	TypeJSONElement
	TypeJSONObject
	TypeJSONArray
)

//  jsonPrimitive

type jsonPrimitive interface {
	timedValue
	getType() TypeOfJSON
	getParent() jsonPrimitive
	getParentAsJSONObject() *jsonObject
	createJSONObject(parent jsonPrimitive, value interface{}, ts *model.Timestamp) *jsonObject
	createJSONArray(parent jsonPrimitive, value interface{}, ts *model.Timestamp) *jsonArray
}

type jsonPrimitiveImpl struct {
	parent jsonPrimitive
}

func (its *jsonPrimitiveImpl) getValue() types.JSONValue {
	panic("should be overridden")
}

func (its *jsonPrimitiveImpl) setValue(v types.JSONValue) {
	panic("should be overridden")
}

func (its *jsonPrimitiveImpl) getTime() *model.Timestamp {
	panic("should be overridden")
}

func (its *jsonPrimitiveImpl) getType() TypeOfJSON {
	return TypeJSONPrimitive
}

func (its *jsonPrimitiveImpl) getParent() jsonPrimitive {
	return its.parent
}

func (its *jsonPrimitiveImpl) getParentAsJSONObject() *jsonObject {
	return its.parent.(*jsonObject)
}

func (its *jsonPrimitiveImpl) String() string {
	return fmt.Sprintf("%x", &its.parent)
}

func (its *jsonPrimitiveImpl) createJSONArray(parent jsonPrimitive, value interface{}, ts *model.Timestamp) *jsonArray {
	ja := newJSONArray(parent, ts.NextDeliminator())
	target := reflect.ValueOf(value)
	var appendValues []timedValue
	for i := 0; i < target.Len(); i++ {
		field := target.Index(i)
		switch field.Kind() {
		case reflect.Slice, reflect.Array:
			ja := its.createJSONArray(ja, field.Interface(), ts)
			appendValues = append(appendValues, ja)
		case reflect.Struct:
			childJO := its.createJSONObject(ja, field.Interface(), ts)
			appendValues = append(appendValues, childJO)
		case reflect.Ptr:
			val := field.Elem()
			its.createJSONArray(parent, val.Interface(), ts)
		default:
			element := newJSONElement(ja, types.ConvertToJSONSupportedValue(field.Interface()), ts.NextDeliminator())
			appendValues = append(appendValues, element)
		}
	}
	if appendValues != nil {
		ja.insertLocalWithTimedValue(0, appendValues...)
	}
	// log.Logger.Infof("%v", ja.String())
	return ja
}

func (its *jsonPrimitiveImpl) createJSONObject(parent jsonPrimitive, value interface{}, ts *model.Timestamp) *jsonObject {
	jo := newJSONObject(parent, ts.NextDeliminator())
	target := reflect.ValueOf(value)
	fields := reflect.TypeOf(value)

	for i := 0; i < target.NumField(); i++ {
		field := target.Field(i)
		switch field.Kind() {
		case reflect.Slice, reflect.Array:
			ja := its.createJSONArray(jo, field.Interface(), ts)
			jo.putCommonWithTimedValue(fields.Field(i).Name, ja)
		case reflect.Struct:
			childJO := its.createJSONObject(jo, field.Interface(), ts)
			jo.putCommonWithTimedValue(fields.Field(i).Name, childJO)
		case reflect.Ptr:
			val := field.Elem()
			its.createJSONObject(parent, val.Interface(), ts)
		default:
			element := newJSONElement(jo, types.ConvertToJSONSupportedValue(field.Interface()), ts.NextDeliminator())
			jo.putCommonWithTimedValue(fields.Field(i).Name, element)
		}
	}
	return jo
}

//  jsonElement

type jsonElement struct {
	jsonPrimitive
	timedValue
}

func newJSONElement(parent jsonPrimitive, value interface{}, ts *model.Timestamp) *jsonElement {
	return &jsonElement{
		jsonPrimitive: &jsonPrimitiveImpl{
			parent: parent,
		},
		timedValue: &timedValueImpl{
			V: value,
			T: ts,
		},
	}
}

func (its *jsonElement) getValue() types.JSONValue {
	return its.timedValue.getValue()
}

func (its *jsonElement) getTime() *model.Timestamp {
	return its.timedValue.getTime()
}

func (its *jsonElement) getType() TypeOfJSON {
	return TypeJSONElement
}

func (its *jsonElement) setValue(v types.JSONValue) {
	panic("not used")
}

func (its *jsonElement) String() string {
	parent := its.getParent()
	parentTS := "nil"
	if parent != nil {
		parentTS = parent.getTime().ToString()
	}
	return fmt.Sprintf("JE(P%v)[T%v|%v]", parentTS, its.getTime().ToString(), its.getValue())
}

//  jsonObject

type jsonObject struct {
	jsonPrimitive
	T *model.Timestamp
	*hashMapSnapshot
}

func newJSONObject(parent jsonPrimitive, ts *model.Timestamp) *jsonObject {
	return &jsonObject{
		T: ts,
		jsonPrimitive: &jsonPrimitiveImpl{
			parent: parent,
		},
		hashMapSnapshot: newHashMapSnapshot(),
	}
}

func (its *jsonObject) getTime() *model.Timestamp {
	return its.T
}

func (its *jsonObject) getValue() types.JSONValue {
	return its.hashMapSnapshot
}

func (its *jsonObject) getType() TypeOfJSON {
	return TypeJSONObject
}

func (its *jsonObject) put(key string, value interface{}, ts *model.Timestamp) {
	rt := reflect.ValueOf(value)
	switch rt.Kind() {
	case reflect.Slice, reflect.Array:
		ja := its.createJSONArray(its, value, ts)
		its.putCommonWithTimedValue(key, ja)
	case reflect.Struct:
		jo := its.createJSONObject(its, value, ts)
		its.putCommonWithTimedValue(key, jo)
	case reflect.Ptr:
		val := rt.Elem()
		log.Logger.Infof("%+v", val.Interface())
		its.put(key, val.Interface(), ts)
	default:
		je := newJSONElement(its, types.ConvertToJSONSupportedValue(value), ts.NextDeliminator())
		its.putCommonWithTimedValue(key, je)
	}
}

func (its *jsonObject) getAsJSONElement(key string) *jsonElement {
	value := its.get(key)
	return value.(*jsonElement)
}

func (its *jsonObject) getAsJSONObject(key string) *jsonObject {
	value := its.get(key)
	return value.(*jsonObject)
}

func (its *jsonObject) getAsJSONArray(key string) *jsonArray {
	value := its.get(key)
	return value.(*jsonArray)
}

func (its *jsonObject) String() string {
	parent := its.getParent()
	parentTS := "nil"
	if parent != nil {
		parentTS = parent.getTime().ToString()
	}
	return fmt.Sprintf("JO(%v)[T%v|V%v]", parentTS, its.T.ToString(), its.hashMapSnapshot.String())
}

func (its *jsonObject) GetAsJSON() interface{} {
	m := make(map[string]interface{})
	for k, v := range its.Map {
		if v != nil {
			switch cast := v.(type) {
			case *jsonObject:
				log.Logger.Infof("%v:%v", k, cast)
				m[k] = cast.GetAsJSON()
			case *jsonElement:
				m[k] = v.getValue()
			case *jsonArray:
				m[k] = cast.GetAsJSON()
			}
		}
	}
	return m
}

//  jsonArray

type jsonArray struct {
	jsonPrimitive
	T *model.Timestamp
	*listSnapshot
}

func newJSONArray(parent jsonPrimitive, ts *model.Timestamp) *jsonArray {
	return &jsonArray{
		T: ts,
		jsonPrimitive: &jsonPrimitiveImpl{
			parent: parent,
		},
		listSnapshot: newListSnapshot(),
	}
}

func (its *jsonArray) insertLocal(pos int32, ts *model.Timestamp, values ...interface{}) {
	var tvs []timedValue
	for _, v := range values {
		rt := reflect.ValueOf(v)
		switch rt.Kind() {
		case reflect.Slice, reflect.Array:
			ja := its.createJSONArray(its, v, ts)
			tvs = append(tvs, ja)
		case reflect.Struct:
			jo := its.createJSONObject(its, v, ts)
			tvs = append(tvs, jo)
		case reflect.Ptr:
			ptrVal := rt.Elem()
			its.insertLocal(pos, ts, ptrVal)
		default:
			je := newJSONElement(its, types.ConvertToJSONSupportedValue(v), ts.NextDeliminator())
			tvs = append(tvs, je)
		}
	}
	its.listSnapshot.insertLocalWithTimedValue(pos, tvs...)

}

func (its *jsonArray) getAsJSONElement(pos int32) (*jsonElement, error) {
	val, err := its.listSnapshot.getTimedValue(pos)
	if err != nil {
		return nil, log.OrtooError(err)
	}
	return val.(*jsonElement), nil
}

func (its *jsonArray) getTime() *model.Timestamp {
	return its.T
}

func (its *jsonArray) getValue() types.JSONValue {
	return its.listSnapshot
}

func (its *jsonArray) getType() TypeOfJSON {
	return TypeJSONArray
}

func (its *jsonArray) setValue(v types.JSONValue) {
	panic("not used")
}

func (its *jsonArray) String() string {
	parent := its.getParent()
	parentTS := "nil"
	if parent != nil {
		parentTS = parent.getTime().ToString()
	}
	return fmt.Sprintf("JA(%v)[T%v|V%v", parentTS, its.T.ToString(), its.listSnapshot.String())
}

//
// type jsonSnapshot struct {
// 	Map  map[string]*element
// 	Size int
// }
//
// type element struct {
// }
//
// // func newJSONObject(value interface{}, ts *model.Timestamp) *jsonObject {
// // 	target := reflect.ValueOf(value)
// // 	elements := target.Elem()
// // 	for i := 0; i < elements.NumField(); i++ {
// // 		mValue := elements.Field(i)
// // 		newHashMapSnapshot()
// // 	}
// // 	return nil
// // }
//
// func newJSONArray() *jsonArray {
// 	return &jsonArray{
// 		listSnapshot: *newListSnapshot(),
// 	}
// }
//
// type jsonArray struct {
// 	listSnapshot
// }
//
// func newJSONSnapshot() *jsonSnapshot {
// 	return &jsonSnapshot{
// 		Map:  make(map[string]*element),
// 		Size: 0,
// 	}
// }
//
// func (its *jsonSnapshot) convertJSONType(value interface{}) interface{} {
// 	rt := reflect.TypeOf(value)
// 	switch rt.Kind() {
// 	case reflect.Slice, reflect.Array:
//
// 	case reflect.Struct:
// 	case reflect.Ptr:
// 		return its.convertJSONType(reflect.Indirect(reflect.ValueOf(value)).Interface())
// 	default:
// 		return types.ConvertToJSONSupportedValue(value)
// 	}
// 	return nil
// }
//
// func (its *jsonSnapshot) createJSONArray(value interface{}, ts *model.Timestamp) *listSnapshot {
// 	jsonArray := newListSnapshot()
// 	var appendValues []interface{}
// 	target := reflect.ValueOf(value)
// 	// fields := reflect.TypeOf(value)
// 	for i := 0; i < target.Len(); i++ {
// 		mValue := target.Index(i)
// 		switch mValue.Kind() {
// 		case reflect.Slice, reflect.Array:
// 			array := its.createJSONArray(mValue.Interface(), ts)
// 			appendValues = append(appendValues, array)
// 		case reflect.Struct:
// 			// object := its.createJSONObject(mValue.Interface(), ts)
// 			// appendValues = append(appendValues, object)
// 		case reflect.Ptr:
// 		default:
// 			appendValues = append(appendValues, types.ConvertToJSONSupportedValue(mValue.Interface()))
// 		}
// 		log.Logger.Infof("%v", mValue.Interface())
// 	}
// 	if appendValues != nil {
// 		jsonArray.appendLocal(ts, appendValues...)
// 	}
// 	log.Logger.Infof("%v", jsonArray.String())
// 	return nil
// }
//
// func (its *jsonSnapshot) putLocal(key string, value interface{}, ts *model.Timestamp) (interface{}, error) {
// 	log.Logger.Infof("PUT value: %v", reflect.TypeOf(value))
// 	rt := reflect.TypeOf(value)
// 	switch rt.Kind() {
// 	case reflect.Slice, reflect.Array:
// 		_ = its.createJSONArray(value, ts)
// 		fmt.Println("is a slice/array with element type", rt.Elem())
// 	case reflect.Struct:
// 		// jsonObject := its.createJSONObject(value, ts)
// 		// log.Logger.Infof("jsonObject: %v", jsonObject)
// 	case reflect.Ptr:
// 		switch rt.Elem().Kind() {
// 		case reflect.Slice, reflect.Array:
// 		case reflect.Struct:
//
// 		}
// 		log.Logger.Infof("%v", rt.Elem())
//
// 	default:
// 		fmt.Println("is something else entirely")
// 	}
// 	return nil, nil
// }
//
// func (its *jsonSnapshot) removeLocal(key string) interface{} {
// 	return nil
// }
//
// func (its *jsonSnapshot) updateLocal(key string) interface{} {
// 	return nil
// }
//
// func (its *jsonSnapshot) getAsJSONObject() *jsonObject {
// 	return nil
// }
//
// func (its *jsonSnapshot) getAsJSONArray() *jsonArray {
// 	return nil
// }
//
// func (its *jsonSnapshot) getAsJSONElement() interface{} {
// 	return nil
// }
