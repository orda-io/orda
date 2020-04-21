package ortoo

import (
	"fmt"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/ortoo/model"
	"github.com/knowhunger/ortoo/ortoo/types"
	"reflect"
)

// ////////////////////////////////////////////////////////////////
//  jsonSnapshot
// ////////////////////////////////////////////////////////////////

type typeOfJSON int

const (
	typeJSONPrimitive typeOfJSON = iota
	typeJSONElement
	typeJSONObject
	typeJSONArray
)

//  jsonPrimitive

type jsonPrimitive interface {
	timedValue
	getType() typeOfJSON
	getParent() jsonPrimitive
	getParentAsJSONObject() *jsonObject
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

func (its *jsonPrimitiveImpl) getType() typeOfJSON {
	return typeJSONPrimitive
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

func (its *jsonElement) getType() typeOfJSON {
	return typeJSONElement
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

func (its *jsonObject) getType() typeOfJSON {
	return typeJSONObject
}

func (its *jsonObject) put(key string, value interface{}, ts *model.Timestamp) {
	rt := reflect.ValueOf(value)
	switch rt.Kind() {
	case reflect.Slice, reflect.Array:
		jsonArray := its.createJSONArray(its, value, ts)
		its.putCommonWithTimedValue(key, jsonArray)
	case reflect.Struct:
		jsonObj := its.createJSONObject(its, value, ts)
		its.putCommonWithTimedValue(key, jsonObj)
	case reflect.Ptr:
		val := rt.Elem()
		log.Logger.Infof("%+v", val.Interface())
		its.put(key, val.Interface(), ts)
	default:
		element := newJSONElement(its, types.ConvertToJSONSupportedValue(value), ts)
		its.putCommonWithTimedValue(key, element)
	}
}

func (its *jsonObject) createJSONArray(parent jsonPrimitive, value interface{}, ts *model.Timestamp) *jsonArray {
	ja := newJSONArray(parent, ts.NextDeliminator())
	target := reflect.ValueOf(value)
	var appendValues []interface{}
	for i := 0; i < target.Len(); i++ {
		mValue := target.Index(i)
		switch mValue.Kind() {
		// case reflect.Slice, reflect.Array:
		// 	child := its.createJSONArray(ja, mValue.Interface(), ts)
		// 	appendValues = append(appendValues, child)
		// case reflect.Struct:
		// 	object := its.createJSONObject(mValue.Interface(), ts)
		// 	appendValues = append(appendValues, object)
		// case reflect.Ptr:
		default:
			element := newJSONElement(ja, types.ConvertToJSONSupportedValue(mValue.Interface()), ts.NextDeliminator())
			appendValues = append(appendValues, element)
		}
		log.Logger.Infof("%v", mValue.Interface())
	}
	if appendValues != nil {
		ja.appendLocal(ts, appendValues...)
	}
	// log.Logger.Infof("%v", ja.String())
	return nil
}

func (its *jsonObject) createJSONObject(parent jsonPrimitive, value interface{}, ts *model.Timestamp) *jsonObject {
	jsonObj := newJSONObject(parent, ts.NextDeliminator())
	target := reflect.ValueOf(value)
	fields := reflect.TypeOf(value)

	for i := 0; i < target.NumField(); i++ {
		mValue := target.Field(i)
		switch mValue.Kind() {
		case reflect.Slice, reflect.Array:
			// array := its.createJSONArray(mValue.Interface(), ts)
			// jsonObject.putCommon(fields.Field(i).Name, array, ts.NextDeliminator())
		case reflect.Struct:
			// object := its.createJSONObject(mValue.Interface(), ts)
			// jsonObject.putCommon(fields.Field(i).Name, object, ts.NextDeliminator())
		case reflect.Ptr:
		default:
			element := newJSONElement(jsonObj, types.ConvertToJSONSupportedValue(mValue.Interface()), ts.NextDeliminator())
			jsonObj.putCommonWithTimedValue(fields.Field(i).Name, element)
		}
	}
	return jsonObj
}

func (its *jsonObject) getAsJSONElement(key string) *jsonElement {
	value := its.get(key)
	return value.(*jsonElement)
}

func (its *jsonObject) getAsJSONObject(key string) *jsonObject {
	value := its.get(key)
	return value.(*jsonObject)
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

func (its *jsonArray) getTime() *model.Timestamp {
	return its.T
}

func (its *jsonArray) getValue() types.JSONValue {
	return its.listSnapshot
}

func (its *jsonArray) getType() typeOfJSON {
	return typeJSONArray
}

func (its *jsonArray) setValue(v types.JSONValue) {
	panic("not used")
}

func (its *jsonArray) String() string {
	return its.listSnapshot.String()
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
