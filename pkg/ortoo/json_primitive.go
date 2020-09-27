package ortoo

import (
	"fmt"
	"github.com/knowhunger/ortoo/pkg/internal/datatypes"
	"github.com/knowhunger/ortoo/pkg/log"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/pkg/types"
	"reflect"
)

// TypeOfJSON is a type to denote the type of JSON.
type TypeOfJSON int

const (
	typeJSONPrimitive TypeOfJSON = iota

	// TypeJSONElement denotes a JSON element type.
	TypeJSONElement
	// TypeJSONObject denotes a JSON object type.
	TypeJSONObject
	// TypeJSONArray denotes a JSON array type.
	TypeJSONArray
)

// jsonType extends timedType
// jsonElement extends jsonType
// jsonObject extends jsonType
// jsonArray extends jsonType

// ////////////////////////////////////
//  jsonType
// ////////////////////////////////////

type jsonType interface {
	timedType
	// getKey() *model.Timestamp
	getType() TypeOfJSON
	getRoot() *jsonCommon
	setRoot(r *jsonObject)
	getParent() jsonType
	setParent(j jsonType)
	getKeyTime() *model.Timestamp
	getDelTime() *model.Timestamp
	getBase() *datatypes.BaseDatatype
	getLogger() *log.OrtooLog
	findJSONArray(ts *model.Timestamp) (j *jsonArray, ok bool)
	findJSONObject(ts *model.Timestamp) (j *jsonObject, ok bool)
	findJSONElement(ts *model.Timestamp) (j *jsonElement, ok bool)
	findJSONPrimitive(ts *model.Timestamp) (j jsonType, ok bool)
	addToNodeMap(j jsonType)
	addToCemetery(j jsonType)
	removeFromNodeMap(j jsonType)
	createJSONType(parent jsonType, v interface{}, ts *model.Timestamp) jsonType
	marshal() *marshaledJSONType
	unmarshal(marshaled *marshaledJSONType, jsonMap map[string]jsonType)
}

type jsonCommon struct {
	root     *jsonObject
	base     *datatypes.BaseDatatype
	nodeMap  map[string]jsonType // store all jsonPrimitive.K.hash => jsonType
	cemetery map[string]jsonType // store all deleted jsonType
}

// jsonPrimitive should implement timedType, and jsonType.
type jsonPrimitive struct {
	common *jsonCommon
	parent jsonType
	T      *model.Timestamp // a timestamp when this primitive is created. This is immutable.
	D      *model.Timestamp // if D is not nil, it is tombstone.
}

// ///////////////////// methods of timedType ///////////////////////////////////

func (its *jsonPrimitive) getTime() *model.Timestamp {
	if its.D != nil {
		return its.D
	}
	return its.T
}

func (its *jsonPrimitive) setTime(ts *model.Timestamp) {
	// since T is immutable and D is used for precedence, it updates D.
	its.D = ts
}

func (its *jsonPrimitive) isTomb() bool {
	return its.D != nil
}

func (its *jsonPrimitive) makeTomb(ts *model.Timestamp) bool {
	if its.D != nil { // Already deleted. This condition meets when another remote delete operation arrives to tombstone.
		if its.D.Compare(ts) > 0 { // if current deletion is older, then ignored.
			return false
		}
		if tomb, ok := its.common.cemetery[its.D.Hash()]; ok {
			delete(its.common.cemetery, its.D.Hash())
			its.common.cemetery[ts.Hash()] = tomb
		}
	}
	its.D = ts
	return true
}

func (its *jsonPrimitive) getValue() types.JSONValue {
	panic("should be overridden")
}

func (its *jsonPrimitive) setValue(v types.JSONValue) {
	panic("should be overridden")
}

// ///////////////////// methods of jsonType ///////////////////////////////////////

func (its *jsonPrimitive) getKeyTime() *model.Timestamp {
	return its.T
}

func (its *jsonPrimitive) getType() TypeOfJSON {
	return typeJSONPrimitive
}

func (its *jsonPrimitive) getBase() *datatypes.BaseDatatype {
	return its.common.base
}

// func (its *jsonPrimitive) makeTombAsChild(ts *model.Timestamp) bool {
// 	if !its.isTomb() {
// 		// its.P = ts
// 		its.deleted = true
// 		return true
// 	}
// 	return false
// }

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
	its.getRoot().nodeMap[primitive.getKeyTime().Hash()] = primitive
}

func (its *jsonPrimitive) removeFromNodeMap(primitive jsonType) {
	delete(its.getRoot().nodeMap, primitive.getKeyTime().Hash())
}

func (its *jsonPrimitive) getDelTime() *model.Timestamp {
	return its.D
}

func (its *jsonPrimitive) addToCemetery(primitive jsonType) {
	its.common.cemetery[primitive.getDelTime().Hash()] = primitive
}

func (its *jsonPrimitive) getRoot() *jsonCommon {
	return its.common
}

func (its *jsonPrimitive) setRoot(r *jsonObject) {
	its.common.root = r
	its.common.nodeMap[r.getKeyTime().Hash()] = r
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

func (its *jsonPrimitive) createJSONTypeFromReflectValue(parent jsonType, rv reflect.Value, ts *model.Timestamp) jsonType {
	switch rv.Kind() {
	case reflect.Struct, reflect.Map:
		return its.createJSONObject(parent, rv.Interface(), ts)
	case reflect.Slice, reflect.Array:
		return its.createJSONArray(parent, rv.Interface(), ts)
	case reflect.Ptr:
		ptrVal := rv.Elem()
		return its.createJSONTypeFromReflectValue(parent, ptrVal, ts)
	default:
		return newJSONElement(parent, types.ConvertToJSONSupportedValue(rv.Interface()), ts.GetAndNextDelimiter())
	}
}

func (its *jsonPrimitive) createJSONType(parent jsonType, v interface{}, ts *model.Timestamp) jsonType {
	rv := reflect.ValueOf(v)
	return its.createJSONTypeFromReflectValue(parent, rv, ts)
}

func (its *jsonPrimitive) createJSONArray(parent jsonType, value interface{}, ts *model.Timestamp) *jsonArray {
	ja := newJSONArray(its.getBase(), parent, ts.GetAndNextDelimiter())
	var appendValues []timedType
	elements := reflect.ValueOf(value)
	for i := 0; i < elements.Len(); i++ {
		rv := elements.Index(i)
		jt := its.createJSONTypeFromReflectValue(ja, rv, ts)
		appendValues = append(appendValues, jt)
	}
	if appendValues != nil {
		_, _, err := ja.insertLocalWithTimedTypes(0, appendValues...)
		if err != nil { // this cannot happen when inserted at position 0.
			_ = log.OrtooError(err)
		}
		for _, v := range appendValues {
			its.addToNodeMap(v.(jsonType))
		}
	}
	return ja
}

func (its *jsonPrimitive) createJSONObject(parent jsonType, value interface{}, ts *model.Timestamp) *jsonObject {
	jo := newJSONObject(its.getBase(), parent, ts.GetAndNextDelimiter())
	target := reflect.ValueOf(value)
	fields := reflect.TypeOf(value)

	if target.Kind() == reflect.Map {
		mapValue := value.(map[string]interface{})
		for k, v := range mapValue {
			val := reflect.ValueOf(v)
			its.addValueToJSONObject(jo, k, val, ts)
		}
	} else { // reflect.Struct
		for i := 0; i < target.NumField(); i++ {
			value := target.Field(i)
			its.addValueToJSONObject(jo, fields.Field(i).Name, value, ts)
		}
	}

	return jo
}

func (its *jsonPrimitive) addValueToJSONObject(jo *jsonObject, key string, value reflect.Value, ts *model.Timestamp) {
	jt := its.createJSONTypeFromReflectValue(jo, value, ts)
	jo.putCommonWithTimedType(key, jt)
	its.addToNodeMap(jt)
}
