package ortoo

import (
	"fmt"
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/iface"
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

type jsonTypeSnapshot interface {
	PutCommonInObject(
		parent *model.Timestamp,
		key string,
		value interface{},
		ts *model.Timestamp,
	) (jsonType, errors.OrtooError)
	DeleteCommonInObject(
		parent *model.Timestamp,
		key string,
		ts *model.Timestamp,
		isLocal bool,
	) (jsonType, errors.OrtooError)
	InsertLocalInArray(
		parent *model.Timestamp,
		pos int,
		ts *model.Timestamp,
		values ...interface{},
	) (*model.Timestamp, jsonType, errors.OrtooError)
	InsertRemoteInArray(
		parent *model.Timestamp,
		target *model.Timestamp,
		ts *model.Timestamp,
		values ...interface{},
	) (jsonType, errors.OrtooError)
	UpdateLocalInArray(
		parent *model.Timestamp,
		pos int,
		ts *model.Timestamp,
		values ...interface{},
	) ([]*model.Timestamp, []jsonType, errors.OrtooError)
	UpdateRemoteInArray(
		parent *model.Timestamp,
		ts *model.Timestamp,
		targets []*model.Timestamp,
		values []interface{},
	) ([]jsonType, errors.OrtooError)
	DeleteLocalInArray(
		parent *model.Timestamp,
		pos, numOfNodes int,
		ts *model.Timestamp,
	) ([]*model.Timestamp, []jsonType, errors.OrtooError)

	DeleteRemoteInArray(
		parent *model.Timestamp,
		ts *model.Timestamp,
		targets []*model.Timestamp,
	) ([]jsonType, errors.OrtooError)
}

// ////////////////////////////////////
//  jsonType
// ////////////////////////////////////

type jsonType interface {
	timedType
	iface.Snapshot
	jsonTypeSnapshot
	getType() TypeOfJSON
	getCommon() *jsonCommon
	getRoot() *jsonObject
	setRoot(r *jsonObject)
	getParent() jsonType
	setParent(j jsonType)
	getCreateTime() *model.Timestamp
	getDeleteTime() *model.Timestamp
	getLogger() *log.OrtooLog
	findJSONArray(ts *model.Timestamp) (j *jsonArray, ok bool)
	findJSONObject(ts *model.Timestamp) (j *jsonObject, ok bool)
	findJSONElement(ts *model.Timestamp) (j *jsonElement, ok bool)
	findJSONType(ts *model.Timestamp) (j jsonType, ok bool)
	addToNodeMap(j jsonType)
	addToCemetery(j jsonType)
	removeFromNodeMap(j jsonType)
	isGarbage() bool
	funeral(j jsonType, ts *model.Timestamp)
	createJSONType(parent jsonType, v interface{}, ts *model.Timestamp) jsonType
	marshal() *marshaledJSONType
	unmarshal(marshaled *marshaledJSONType, assistant *unmarshalAssistant)
	equal(o jsonType) bool
}

type jsonCommon struct {
	root     *jsonObject
	base     base
	nodeMap  map[string]jsonType // store all jsonPrimitive.K.hash => jsonType
	cemetery map[string]jsonType // store all deleted jsonType
}

func (its *jsonCommon) equal(o *jsonCommon) bool {
	if its.base != o.base {
		return false
	}
	for k, v1 := range its.nodeMap {
		v2 := o.nodeMap[k]
		if !v1.equal(v2) {
			its.base.L().Errorf("\n%v\n%v", v1, v2)
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

func (its *jsonCommon) SetBase(base iface.BaseDatatype) {
	its.base = base
	for _, node := range its.nodeMap {
		switch cast := node.(type) {
		case *jsonObject:
			cast.base = base
		case *jsonArray:
			cast.base = base
		}
	}
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

func (its *jsonPrimitive) isGarbage() bool {
	var p jsonType = its
	for p != nil {
		if p.isTomb() {
			return true
		}
		p = p.getParent()
	}
	return false
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

func (its *jsonPrimitive) getLogger() *log.OrtooLog {
	return its.common.base.L()
}

func (its *jsonPrimitive) findJSONType(ts *model.Timestamp) (j jsonType, ok bool) {
	node, ok := its.getCommon().nodeMap[ts.Hash()]
	return node, ok
}

func (its *jsonPrimitive) findJSONElement(ts *model.Timestamp) (j *jsonElement, ok bool) {
	if node, ok := its.getCommon().nodeMap[ts.Hash()]; ok {
		if j, ok2 := node.(*jsonElement); ok2 {
			return j, ok2
		}
	}
	return nil, false
}

func (its *jsonPrimitive) findJSONObject(ts *model.Timestamp) (json *jsonObject, ok bool) {
	if node, ok := its.getCommon().nodeMap[ts.Hash()]; ok {
		if j, ok2 := node.(*jsonObject); ok2 {
			return j, ok2
		}
	}
	return nil, false
}

func (its *jsonPrimitive) findJSONArray(ts *model.Timestamp) (json *jsonArray, ok bool) {
	if node, ok := its.getCommon().nodeMap[ts.Hash()]; ok {
		if j, ok2 := node.(*jsonArray); ok2 {
			return j, ok2
		}
	}
	return nil, false
}

func (its *jsonPrimitive) addToNodeMap(primitive jsonType) {
	its.getCommon().nodeMap[primitive.getCreateTime().Hash()] = primitive
}

func (its *jsonPrimitive) removeFromNodeMap(primitive jsonType) {
	delete(its.getCommon().nodeMap, primitive.getCreateTime().Hash())
}

func (its *jsonPrimitive) getDeleteTime() *model.Timestamp {
	return its.D
}

func (its *jsonPrimitive) addToCemetery(primitive jsonType) {
	its.common.cemetery[primitive.getDeleteTime().Hash()] = primitive
}

func (its *jsonPrimitive) getCommon() *jsonCommon {
	return its.common
}

func (its *jsonPrimitive) getRoot() *jsonObject {
	return its.common.root
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
	ja := newJSONArray(its.GetBase(), parent, ts.GetAndNextDelimiter())
	var appendValues []timedType
	elements := reflect.ValueOf(value)
	for i := 0; i < elements.Len(); i++ {
		rv := elements.Index(i)
		jt := its.createJSONTypeFromReflectValue(ja, rv, ts)
		appendValues = append(appendValues, jt)
	}
	if appendValues != nil {
		_, _ = ja.insertLocalWithTimedTypes(0, appendValues...)
		for _, v := range appendValues {
			its.addToNodeMap(v.(jsonType))
		}
	}
	return ja
}

func (its *jsonPrimitive) createJSONObject(parent jsonType, value interface{}, ts *model.Timestamp) *jsonObject {
	jo := newJSONObject(its.GetBase(), parent, ts.GetAndNextDelimiter())
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

func (its *jsonPrimitive) PutCommonInObject(
	parent *model.Timestamp,
	key string,
	value interface{},
	ts *model.Timestamp,
) (jsonType, errors.OrtooError) {
	if parentObj, ok := its.findJSONObject(parent); ok {
		return parentObj.putCommon(key, value, ts), nil
	}
	return nil, errors.DatatypeInvalidParent.New(its.getLogger(), parent.ToString())
}

func (its *jsonPrimitive) DeleteCommonInObject(
	parent *model.Timestamp,
	key string,
	ts *model.Timestamp,
	isLocal bool,
) (jsonType, errors.OrtooError) {
	if parentObj, ok := its.findJSONObject(parent); ok {
		return parentObj.deleteCommonInObject(key, ts, isLocal)
	}
	return nil, errors.DatatypeInvalidParent.New(its.getLogger(), parent.ToString())
}

// InsertLocalInArray inserts values locally and returns the timestamp of target
func (its *jsonPrimitive) InsertLocalInArray(
	parent *model.Timestamp,
	pos int,
	ts *model.Timestamp,
	values ...interface{},
) (
	*model.Timestamp, // the timestamp of target
	jsonType, // parent Array
	errors.OrtooError, // error
) {
	if parentArray, ok := its.findJSONArray(parent); ok {
		target, _, err := parentArray.insertCommon(pos, nil, ts, values...)
		return target, parentArray, err
	}
	return nil, nil, errors.DatatypeInvalidParent.New(its.getLogger(), parent.ToString())
}

func (its *jsonPrimitive) InsertRemoteInArray(
	parent *model.Timestamp,
	target *model.Timestamp,
	ts *model.Timestamp,
	values ...interface{},
) (jsonType, errors.OrtooError) {
	if parentArray, ok := its.findJSONArray(parent); ok {
		_, _, err := parentArray.insertCommon(-1, target, ts, values...)
		return parentArray, err
	}
	return nil, errors.DatatypeInvalidParent.New(its.getLogger(), parent.ToString())
}

func (its *jsonPrimitive) UpdateLocalInArray(
	parent *model.Timestamp,
	pos int,
	ts *model.Timestamp,
	values ...interface{},
) ([]*model.Timestamp, []jsonType, errors.OrtooError) {
	if parentArray, ok := its.findJSONArray(parent); ok {
		return parentArray.updateLocal(pos, ts, values...)
	}
	return nil, nil, errors.DatatypeInvalidParent.New(its.getLogger(), parent.ToString())
}

func (its *jsonPrimitive) UpdateRemoteInArray(
	parent *model.Timestamp,
	ts *model.Timestamp,
	targets []*model.Timestamp,
	values []interface{},
) ([]jsonType, errors.OrtooError) {
	if parentArray, ok := its.findJSONArray(parent); ok {
		return parentArray.updateRemote(ts, targets, values)
	}
	return nil, errors.DatatypeInvalidParent.New(its.getLogger(), parent.ToString())
}

func (its *jsonPrimitive) DeleteLocalInArray(
	parent *model.Timestamp,
	pos, numOfNodes int,
	ts *model.Timestamp,
) ([]*model.Timestamp, []jsonType, errors.OrtooError) {
	if parentArray, ok := its.findJSONArray(parent); ok {
		t, j := parentArray.deleteLocal(pos, numOfNodes, ts)
		return t, j, nil
	}
	return nil, nil, errors.DatatypeInvalidParent.New(its.getLogger(), parent.ToString())
}

func (its *jsonPrimitive) DeleteRemoteInArray(
	parent *model.Timestamp,
	ts *model.Timestamp,
	targets []*model.Timestamp,
) ([]jsonType, errors.OrtooError) {
	if parentArray, ok := its.findJSONArray(parent); ok {
		return parentArray.deleteRemote(targets, ts)
	}
	return nil, errors.DatatypeInvalidParent.New(its.getLogger(), parent.ToString())
}

// ///////////////////// methods of iface.Snapshot ///////////////////////////////////////

func (its *jsonPrimitive) SetBase(base iface.BaseDatatype) {
	its.common.SetBase(base)
}

func (its *jsonPrimitive) GetBase() iface.BaseDatatype {
	return its.common.base
}

func (its *jsonPrimitive) CloneSnapshot() iface.Snapshot {
	panic("Implement me")
}

func (its *jsonPrimitive) GetAsJSONCompatible() interface{} {
	return its.getValue()
}
