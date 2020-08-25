package ortoo

import (
	"fmt"
	"github.com/knowhunger/ortoo/pkg/internal/datatypes"
	"github.com/knowhunger/ortoo/pkg/log"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/pkg/types"
	"reflect"
)

type TypeOfJSON int

const (
	typeJSONPrimitive TypeOfJSON = iota
	TypeJSONElement
	TypeJSONObject
	TypeJSONArray
)

// jsonType extends precededType
// jsonElement extends jsonType
// jsonObject extends jsonType
// jsonArray extends jsonType

// ////////////////////////////////////
//  jsonType
// ////////////////////////////////////

type jsonType interface {
	precededType
	// getKey() *model.Timestamp
	getType() TypeOfJSON
	getRoot() *jsonCommon
	setRoot(r *jsonObject)
	getParent() jsonType
	setParent(j jsonType)
	getBase() *datatypes.BaseDatatype
	getLogger() *log.OrtooLog
	// makeTombAsChild() makes tomb when it is not a tomb
	makeTombAsChild(ts *model.Timestamp) bool
	findJSONArray(ts *model.Timestamp) (j *jsonArray, ok bool)
	findJSONObject(ts *model.Timestamp) (j *jsonObject, ok bool)
	findJSONElement(ts *model.Timestamp) (j *jsonElement, ok bool)
	findJSONPrimitive(ts *model.Timestamp) (j jsonType, ok bool)
	addToNodeMap(j jsonType)
	addToCemetery(j jsonType)
	createJSONObject(parent jsonType, value interface{}, ts *model.Timestamp) *jsonObject
	createJSONArray(parent jsonType, value interface{}, ts *model.Timestamp) *jsonArray
	marshal() *marshaledJSONType
	unmarshal(marshaled *marshaledJSONType, jsonMap map[string]jsonType)
}

type jsonCommon struct {
	root     *jsonObject
	base     *datatypes.BaseDatatype
	nodeMap  map[string]jsonType // store all jsonPrimitive.K.hash => jsonType
	cemetery map[string]jsonType // store all jsonPrimitive.K.hash => deleted jsonType
}

type jsonPrimitive struct {
	common  *jsonCommon
	parent  jsonType
	deleted bool
	K       *model.Timestamp // used for key that is immutable and used in the common
	P       *model.Timestamp // used for precedence; for example makeTomb
}

func (its *jsonPrimitive) unmarshal(marshaled *marshaledJSONType, jsonMap map[string]jsonType) {
	// do nothing
}

func (its *jsonPrimitive) getType() TypeOfJSON {
	return typeJSONPrimitive
}

func (its *jsonPrimitive) getBase() *datatypes.BaseDatatype {
	return its.common.base
}

func (its *jsonPrimitive) makeTombAsChild(ts *model.Timestamp) bool {
	if !its.isTomb() {
		its.P = ts
		its.deleted = true
		return true
	}
	return false
}

func (its *jsonPrimitive) isTomb() bool {
	return its.deleted
}

func (its *jsonPrimitive) makeTomb(ts *model.Timestamp) bool {
	if its.deleted { // already deleted
		if its.P.Compare(ts) > 0 { // if current deletion is older, then ignored.
			log.Logger.Infof("fail to makeTomb() of jsonPrimitive:%v", its.K.ToString())
			return false
		}
	}
	its.P = ts
	its.deleted = true
	return true
}

func (its *jsonPrimitive) getKey() *model.Timestamp {
	return its.K
}

func (its *jsonPrimitive) getTime() *model.Timestamp {
	if its.P == nil {
		return its.K
	}
	return its.P
}

func (its *jsonPrimitive) setTime(ts *model.Timestamp) {
	its.K = ts
}

func (its *jsonPrimitive) getPrecedence() *model.Timestamp {
	return its.P
}

func (its *jsonPrimitive) setPrecedence(ts *model.Timestamp) {
	its.P = ts
}

func (its *jsonPrimitive) getLogger() *log.OrtooLog {
	return its.common.base.Logger
}

func (its *jsonPrimitive) findJSONPrimitive(ts *model.Timestamp) (j jsonType, ok bool) {
	node, ok := its.getRoot().nodeMap[ts.Hash()]
	return node, ok
}

func (its *jsonPrimitive) findJSONElement(ts *model.Timestamp) (j *jsonElement, ok bool) {
	if node, ok := its.getRoot().nodeMap[ts.Hash()]; ok {
		if j, ok2 := node.(*jsonElement); ok2 {
			return j, ok2
		}
	}
	return nil, false
}

func (its *jsonPrimitive) findJSONObject(ts *model.Timestamp) (json *jsonObject, ok bool) {
	if node, ok := its.getRoot().nodeMap[ts.Hash()]; ok {
		if j, ok2 := node.(*jsonObject); ok2 {
			return j, ok2
		}
	}
	return nil, false
}

func (its *jsonPrimitive) findJSONArray(ts *model.Timestamp) (json *jsonArray, ok bool) {
	if node, ok := its.getRoot().nodeMap[ts.Hash()]; ok {
		if j, ok2 := node.(*jsonArray); ok2 {
			return j, ok2
		}
	}
	return nil, false
}

func (its *jsonPrimitive) addToNodeMap(primitive jsonType) {
	its.getRoot().nodeMap[primitive.getKey().Hash()] = primitive
}

func (its *jsonPrimitive) addToCemetery(primitive jsonType) {
	its.getRoot().cemetery[primitive.getKey().Hash()] = primitive
}

func (its *jsonPrimitive) getValue() types.JSONValue {
	panic("should be overridden")
}

func (its *jsonPrimitive) setValue(v types.JSONValue) {
	panic("should be overridden")
}

func (its *jsonPrimitive) getRoot() *jsonCommon {
	return its.common
}

func (its *jsonPrimitive) setRoot(r *jsonObject) {
	its.common.root = r
	its.common.nodeMap[r.getTime().Hash()] = r
}

func (its *jsonPrimitive) getParent() jsonType {
	return its.parent
}

func (its *jsonPrimitive) setParent(j jsonType) {
	its.parent = j
}

func (its *jsonPrimitive) String() string {
	return fmt.Sprintf("%x", &its.parent)
}

func (its *jsonPrimitive) createJSONArray(parent jsonType, value interface{}, ts *model.Timestamp) *jsonArray {
	ja := newJSONArray(its.getBase(), parent, ts.NextDeliminator())
	target := reflect.ValueOf(value)
	var appendValues []precededType
	for i := 0; i < target.Len(); i++ {
		field := target.Index(i)
		switch field.Kind() {
		case reflect.Slice, reflect.Array:
			ja := its.createJSONArray(ja, field.Interface(), ts)
			appendValues = append(appendValues, ja)
		case reflect.Struct, reflect.Map:
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
		ja.insertLocalWithPrecededTypes(0, appendValues...)
		for _, v := range appendValues {
			its.addToNodeMap(v.(jsonType))
		}
	}
	// log.Logger.Infof("%v", ja.String())
	return ja
}

func (its *jsonPrimitive) createJSONObject(parent jsonType, value interface{}, ts *model.Timestamp) *jsonObject {
	jo := newJSONObject(its.getBase(), parent, ts.NextDeliminator())
	target := reflect.ValueOf(value)
	fields := reflect.TypeOf(value)

	if target.Kind() == reflect.Map {
		mapValue := value.(map[string]interface{})
		for k, v := range mapValue {
			val := reflect.ValueOf(v)
			its.addValueToJSONObject(jo, k, val, ts)
		}
	} else {
		for i := 0; i < target.NumField(); i++ {
			value := target.Field(i)
			its.addValueToJSONObject(jo, fields.Field(i).Name, value, ts)
		}
	}

	return jo
}

func (its *jsonPrimitive) addValueToJSONObject(jo *jsonObject, key string, value reflect.Value, ts *model.Timestamp) {
	switch value.Kind() {
	case reflect.Slice, reflect.Array:
		ja := its.createJSONArray(jo, value.Interface(), ts)
		jo.putCommonWithTimedValue(key, ja)
		its.addToNodeMap(ja)
	case reflect.Struct, reflect.Map:
		childJO := its.createJSONObject(jo, value.Interface(), ts)
		jo.putCommonWithTimedValue(key, childJO)
		its.addToNodeMap(childJO)
	case reflect.Ptr:
		val := value.Elem()
		its.createJSONObject(jo, val.Interface(), ts)
	default:
		element := newJSONElement(jo, types.ConvertToJSONSupportedValue(value.Interface()), ts.NextDeliminator())
		jo.putCommonWithTimedValue(key, element)
		its.addToNodeMap(element)
	}
}
