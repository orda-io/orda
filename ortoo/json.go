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

type jsonSnapshot struct {
	Map  map[string]*element
	Size int
}

type element struct {
}

func newJSONObject() *jsonObject {
	return &jsonObject{
		hashMapSnapshot: *newHashMapSnapshot(),
	}
}

type jsonObject struct {
	hashMapSnapshot
}

// func newJSONObject(value interface{}, ts *model.Timestamp) *jsonObject {
// 	target := reflect.ValueOf(value)
// 	elements := target.Elem()
// 	for i := 0; i < elements.NumField(); i++ {
// 		mValue := elements.Field(i)
// 		newHashMapSnapshot()
// 	}
// 	return nil
// }

func newJSONArray() *jsonArray {
	return &jsonArray{
		listSnapshot: *newListSnapshot(),
	}
}

type jsonArray struct {
	listSnapshot
}

func newJSONSnapshot() *jsonSnapshot {
	return &jsonSnapshot{
		Map:  make(map[string]*element),
		Size: 0,
	}
}

func (its *jsonSnapshot) convertJSONType(value interface{}) interface{} {
	rt := reflect.TypeOf(value)
	switch rt.Kind() {
	case reflect.Slice, reflect.Array:

	case reflect.Struct:
	case reflect.Ptr:
		return its.convertJSONType(reflect.Indirect(reflect.ValueOf(value)).Interface())
	default:
		return types.ConvertToJSONSupportedType(value)
	}
	return nil
}

func (its *jsonSnapshot) createJSONArray(value interface{}, ts *model.Timestamp) *listSnapshot {
	jsonArray := newListSnapshot()
	var appendValues []interface{}
	target := reflect.ValueOf(value)
	// fields := reflect.TypeOf(value)
	for i := 0; i < target.Len(); i++ {
		mValue := target.Index(i)
		switch mValue.Kind() {
		case reflect.Slice, reflect.Array:
			array := its.createJSONArray(mValue.Interface(), ts)
			appendValues = append(appendValues, array)
		case reflect.Struct:
			object := its.createJSONObject(mValue.Interface(), ts)
			appendValues = append(appendValues, object)
		case reflect.Ptr:
		default:
			appendValues = append(appendValues, types.ConvertToJSONSupportedType(mValue.Interface()))
		}
		log.Logger.Infof("%v", mValue.Interface())
	}
	if appendValues != nil {
		jsonArray.appendLocal(ts, appendValues...)
	}
	log.Logger.Infof("%v", jsonArray.String())
	return nil
}

func (its *jsonSnapshot) createJSONObject(value interface{}, ts *model.Timestamp) *hashMapSnapshot {
	jsonObject := newHashMapSnapshot()
	target := reflect.ValueOf(value)
	fields := reflect.TypeOf(value)

	for i := 0; i < target.NumField(); i++ {
		mValue := target.Field(i)
		switch mValue.Kind() {
		case reflect.Slice, reflect.Array:
			array := its.createJSONArray(mValue.Interface(), ts)
			jsonObject.putCommon(fields.Field(i).Name, array, ts.NextDeliminator())
		case reflect.Struct:
			object := its.createJSONObject(mValue.Interface(), ts)
			jsonObject.putCommon(fields.Field(i).Name, object, ts.NextDeliminator())
		case reflect.Ptr:
		default:
			jsonObject.putCommon(fields.Field(i).Name, types.ConvertToJSONSupportedType(mValue.Interface()), ts.NextDeliminator())
			log.Logger.Infof("Key:%v Type:%v Value:%v", fields.Field(i).Name, mValue.Type(), mValue.Interface())
		}
	}
	return jsonObject
}

func (its *jsonSnapshot) putLocal(key string, value interface{}, ts *model.Timestamp) (interface{}, error) {
	log.Logger.Infof("PUT value: %v", reflect.TypeOf(value))
	rt := reflect.TypeOf(value)
	switch rt.Kind() {
	case reflect.Slice, reflect.Array:
		_ = its.createJSONArray(value, ts)
		fmt.Println("is a slice/array with element type", rt.Elem())
	case reflect.Struct:
		jsonObject := its.createJSONObject(value, ts)
		log.Logger.Infof("jsonObject: %v", jsonObject)
	case reflect.Ptr:
		switch rt.Elem().Kind() {
		case reflect.Slice, reflect.Array:
		case reflect.Struct:

		}
		log.Logger.Infof("%v", rt.Elem())

	default:
		fmt.Println("is something else entirely")
	}
	return nil, nil
}

func (its *jsonSnapshot) removeLocal(key string) interface{} {
	return nil
}

func (its *jsonSnapshot) updateLocal(key string) interface{} {
	return nil
}

func (its *jsonSnapshot) getAsJSONObject() *jsonObject {
	return nil
}

func (its *jsonSnapshot) getAsJSONArray() *jsonArray {
	return nil
}

func (its *jsonSnapshot) getAsJSONElement() interface{} {
	return nil
}
