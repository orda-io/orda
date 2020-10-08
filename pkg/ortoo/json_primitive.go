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

var (
	typeName = map[TypeOfJSON]string{
		TypeJSONElement: "JSONElement",
		TypeJSONObject:  "JSONObject",
		TypeJSONArray:   "JSONArray",
	}
)

// Name returns the name of JSONType
func (its TypeOfJSON) Name() string {
	return typeName[its]
}

// jsonType extends timedType
// jsonElement extends jsonType
// jsonObject extends jsonType
// jsonArray extends jsonType

// ////////////////////////////////////
//  jsonType
// ////////////////////////////////////

type jsonType interface {
	timedType
	getType() TypeOfJSON
	getRoot() *jsonCommon
	setRoot(r *jsonObject)
	getParent() jsonType
	setParent(j jsonType)
	getCreateTime() *model.Timestamp
	getDeleteTime() *model.Timestamp
	getBase() *datatypes.BaseDatatype
	getLogger() *log.OrtooLog
	findJSONArray(ts *model.Timestamp) (j *jsonArray, ok bool)
	findJSONObject(ts *model.Timestamp) (j *jsonObject, ok bool)
	findJSONElement(ts *model.Timestamp) (j *jsonElement, ok bool)
	findJSONType(ts *model.Timestamp) (j jsonType, ok bool)
	addToNodeMap(j jsonType)
	addToCemetery(j jsonType)
	removeFromNodeMap(j jsonType)
	funeral(j jsonType, ts *model.Timestamp)
	createJSONType(parent jsonType, v interface{}, ts *model.Timestamp) jsonType
	marshal() *marshaledJSONType
	unmarshal(marshaled *marshaledJSONType, assistant *unmarshalAssistant)
	equal(o jsonType) bool
}

type jsonCommon struct {
	root     *jsonObject
	base     *datatypes.BaseDatatype
	nodeMap  map[string]jsonType // store all jsonPrimitive.K.hash => jsonType
	cemetery map[string]jsonType // store all deleted jsonType
}

func (its *jsonCommon) equal(o *jsonCommon) bool {
	for k, v1 := range its.nodeMap {
		v2 := o.nodeMap[k]
		if !v1.equal(v2) {
			log.Logger.Errorf("\n%v\n%v", v1, v2)
			return false
		}
	}
	if len(its.cemetery) != len(o.cemetery) {
		return false
	}
	for k, v1 := range its.cemetery {
		v2 := o.cemetery[k]
		if v2 == nil {
			return false
		}
		if !(v1.isTomb() && v2.isTomb()) {
			return false
		}
		if v1.getCreateTime().Compare(v2.getCreateTime()) != 0 {
			return false
		}
	}
	return true
}

// jsonPrimitive should implement timedType, and jsonType.
type jsonPrimitive struct {
	common *jsonCommon
	parent jsonType
	C      *model.Timestamp // a timestamp when this primitive is created. This is immutable.
	D      *model.Timestamp // if D is not nil, it is tombstone.
}

// ///////////////////// methods of timedType ///////////////////////////////////

func (its *jsonPrimitive) getTime() *model.Timestamp {
	if its.D != nil {
		return its.D
	}
	return its.C
}

func (its *jsonPrimitive) setTime(ts *model.Timestamp) {
	// since C is immutable and D is used for precedence, it updates D.
	its.D = ts
}

func (its *jsonPrimitive) isTomb() bool {
	return its.D != nil
}

func (its *jsonPrimitive) funeral(j jsonType, ts *model.Timestamp) {
	j.makeTomb(ts)
	if j.getType() == TypeJSONElement {
		its.removeFromNodeMap(j) // jsonElement doesn't need to be accessed, thus is garbage-collected.
	} else {
		its.addToCemetery(j)
	}
}

func (its *jsonPrimitive) makeTomb(ts *model.Timestamp) {
	if its.D != nil { // Only when already tombstone.
		// Since a tombstone is placed in the cemetery based on its.D, it should be adjusted.
		if tomb, ok := its.common.cemetery[its.D.Hash()]; ok {
			delete(its.common.cemetery, its.D.Hash())
			its.common.cemetery[ts.Hash()] = tomb
		}
	}
	its.D = ts
}

func (its *jsonPrimitive) getValue() types.JSONValue {
	panic("should be overridden")
}

func (its *jsonPrimitive) setValue(v types.JSONValue) {
	panic("should be overridden")
}

// ///////////////////// methods of jsonType ///////////////////////////////////////

func (its *jsonPrimitive) getCreateTime() *model.Timestamp {
	return its.C
}

func (its *jsonPrimitive) getType() TypeOfJSON {
	return typeJSONPrimitive
}

func (its *jsonPrimitive) getBase() *datatypes.BaseDatatype {
	return its.common.base
}

func (its *jsonPrimitive) getLogger() *log.OrtooLog {
	return its.common.base.Logger
}

func (its *jsonPrimitive) findJSONType(ts *model.Timestamp) (j jsonType, ok bool) {
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
	its.getRoot().nodeMap[primitive.getCreateTime().Hash()] = primitive
}

func (its *jsonPrimitive) removeFromNodeMap(primitive jsonType) {
	delete(its.getRoot().nodeMap, primitive.getCreateTime().Hash())
}

func (its *jsonPrimitive) getDeleteTime() *model.Timestamp {
	return its.D
}

func (its *jsonPrimitive) addToCemetery(primitive jsonType) {
	its.common.cemetery[primitive.getDeleteTime().Hash()] = primitive
}

func (its *jsonPrimitive) getRoot() *jsonCommon {
	return its.common
}

func (its *jsonPrimitive) setRoot(r *jsonObject) {
	its.common.root = r
	its.common.nodeMap[r.getCreateTime().Hash()] = r
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

func (its *jsonPrimitive) equal(o jsonType) bool {
	if its.getType() != o.getType() {
		return false
	}
	if its.getCreateTime().Compare(o.getCreateTime()) != 0 {
		return false
	}
	if its.getDeleteTime() != nil && o.getDeleteTime() != nil &&
		its.getDeleteTime().Compare(o.getDeleteTime()) != 0 {
		return false
	}
	if (its.getDeleteTime() == nil && o.getDeleteTime() != nil) ||
		(its.getDeleteTime() != nil && o.getDeleteTime() == nil) {
		return false
	}
	if its.getParent() != nil && o.getParent() != nil &&
		its.getParent().getCreateTime().Compare(o.getParent().getCreateTime()) != 0 {
		return false
	}
	if (its.getParent() == nil && o.getParent() != nil) ||
		(its.getParent() != nil && o.getParent() == nil) {
		return false
	}
	return true
}
