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
}

type jsonArray struct {
}

func newJSONSnapshot() *jsonSnapshot {
	return &jsonSnapshot{
		Map:  make(map[string]*element),
		Size: 0,
	}
}

func (its *jsonSnapshot) transformJSONType(value interface{}) interface{} {
	rt := reflect.TypeOf(value)
	switch rt.Kind() {
	case reflect.Slice, reflect.Array:

	case reflect.Struct:
	case reflect.Ptr:
		// var val = *value.(*interface{})
		log.Logger.Infof("%v", rt.Elem().Kind())
		switch rt.Elem().Kind() {
		case reflect.String:
			// rt.Elem().Field(0)
			return types.ConvertToJSONSupportedType(reflect.ValueOf(value).String())
		}

		return types.ConvertToJSONSupportedType(rt.Elem())
	default:
		return types.ConvertToJSONSupportedType(value)
	}
	return nil
}

func (its *jsonSnapshot) putLocal(key string, value interface{}, ts *model.Timestamp) (interface{}, error) {

	log.Logger.Infof("%v", reflect.TypeOf(value))
	rt := reflect.TypeOf(value)
	switch rt.Kind() {
	case reflect.Slice, reflect.Array:
		fmt.Println("is a slice/array with element type", rt.Elem())
	case reflect.Struct:
		fmt.Println("is a struct with element type", rt)
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
