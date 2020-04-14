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

type jsonObject struct {
	Map map[string]*obj
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

type jsonArray struct {
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

func (its *jsonSnapshot) createJSONObject(value interface{}, ts *model.Timestamp) *hashMapSnapshot {
	jsonObject := newHashMapSnapshot()
	target := reflect.ValueOf(value)
	fields := reflect.TypeOf(value)
	// elements := target.Elem()
	for i := 0; i < target.NumField(); i++ {
		mValue := target.Field(i)
		switch mValue.Kind() {
		case reflect.Slice, reflect.Array:
		case reflect.Struct:
			j := its.createJSONObject(mValue.Interface(), ts)
			jsonObject.putCommon(fields.Field(i).Name, j, ts.NextDeliminator())
		case reflect.Ptr:
		default:
			jsonObject.putCommon(fields.Field(i).Name, types.ConvertToJSONSupportedType(mValue.Interface()), ts.NextDeliminator())

			log.Logger.Infof("%v %v %v", fields.Field(i).Name, mValue.Type(), mValue.Interface())
		}
	}
	return jsonObject
}

func (its *jsonSnapshot) putLocal(key string, value interface{}, ts *model.Timestamp) (interface{}, error) {

	log.Logger.Infof("%v", reflect.TypeOf(value))
	rt := reflect.TypeOf(value)
	switch rt.Kind() {
	case reflect.Slice, reflect.Array:
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
