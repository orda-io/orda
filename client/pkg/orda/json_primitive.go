package orda

import (
	"fmt"
	"github.com/orda-io/orda/client/pkg/errors"
	"github.com/orda-io/orda/client/pkg/iface"
	"github.com/orda-io/orda/client/pkg/log"
	"github.com/orda-io/orda/client/pkg/model"
	"github.com/orda-io/orda/client/pkg/types"
	"github.com/orda-io/orda/client/pkg/utils"
	"reflect"
	"strconv"
	"strings"
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
	) (jsonType, errors.OrdaError)
	DeleteCommonInObject(
		parent *model.Timestamp,
		key string,
		ts *model.Timestamp,
		isLocal bool,
	) (jsonType, errors.OrdaError)
	InsertLocalInArray(
		parent *model.Timestamp,
		pos int,
		ts *model.Timestamp,
		values ...interface{},
	) (*model.Timestamp, jsonType, errors.OrdaError)
	InsertRemoteInArray(
		parent *model.Timestamp,
		target *model.Timestamp,
		ts *model.Timestamp,
		values ...interface{},
	) (jsonType, errors.OrdaError)
	UpdateLocalInArray(
		parent *model.Timestamp,
		pos int,
		ts *model.Timestamp,
		values ...interface{},
	) ([]*model.Timestamp, []jsonType, errors.OrdaError)
	UpdateRemoteInArray(
		parent *model.Timestamp,
		ts *model.Timestamp,
		targets []*model.Timestamp,
		values []interface{},
	) ([]jsonType, errors.OrdaError)
	DeleteLocalInArray(
		parent *model.Timestamp,
		pos, numOfNodes int,
		ts *model.Timestamp,
	) ([]*model.Timestamp, []jsonType, errors.OrdaError)
	DeleteRemoteInArray(
		parent *model.Timestamp,
		ts *model.Timestamp,
		targets []*model.Timestamp,
	) ([]jsonType, errors.OrdaError)
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
	getLogger() *log.OrdaLog
	findJSONArray(ts *model.Timestamp) (j *jsonArray, ok bool)
	findJSONObject(ts *model.Timestamp) (j *jsonObject, ok bool)
	findJSONElement(ts *model.Timestamp) (j *jsonElement, ok bool)
	findJSONType(ts *model.Timestamp) (j jsonType, ok bool)
	addToNodeMap(j jsonType)
	addToCemetery(j jsonType)
	removeFromNodeMap(j jsonType)
	getTargetByPaths(paths []string) (jsonType, errors.OrdaError)
	getTargetFromPatch(path string) (jsonType, string, errors.OrdaError)
	isGarbage() bool
	funeral(j jsonType, ts *model.Timestamp)
	createJSONType(parent jsonType, v interface{}, ts *model.Timestamp) jsonType
	marshal() *marshaledJSONType
	unmarshal(marshaled *marshaledJSONType, assistant *unmarshalAssistant)
	equal(o jsonType) bool
}

type jsonCommon struct {
	iface.BaseDatatype
	root     *jsonObject
	NodeMap  map[string]jsonType // store all jsonPrimitive.K.hash => jsonType
	Cemetery map[string]jsonType // store all deleted jsonType
}

func (its *jsonCommon) equal(o *jsonCommon) bool {
	if its.BaseDatatype != o.BaseDatatype {
		return false
	}
	for k, v1 := range its.NodeMap {
		v2 := o.NodeMap[k]
		if !v1.equal(v2) {
			its.L().Errorf("\n%v\n%v", v1, v2)
			return false
		}
	}
	if len(its.Cemetery) != len(o.Cemetery) {
		return false
	}
	for k, v1 := range its.Cemetery {
		v2 := o.Cemetery[k]
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
	its.BaseDatatype = base
	for _, node := range its.NodeMap {
		switch cast := node.(type) {
		case *jsonObject:
			cast.BaseDatatype = base
		case *jsonArray:
			cast.BaseDatatype = base
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

func (its *jsonPrimitive) getTargetByPaths(paths []string) (jsonType, errors.OrdaError) {
	var node jsonType = its.getRoot()
	for _, s := range paths {

		switch node.getType() {
		case TypeJSONElement:
			its.common.L().Errorf("invalid target")
		case TypeJSONObject:
			node = node.(*jsonObject).getAsJSONType(s)
		case TypeJSONArray:
			pos, err := strconv.Atoi(s)
			if err != nil {
				return nil, errors.DatatypeNoTarget.New(its.common.L(), "invalid path:%v from %v", s, strings.Join(paths, "/"))
			}
			node = node.(*jsonArray).getJSONType(pos)
		}

		if node == nil || node.isGarbage() {
			return nil, errors.DatatypeNoTarget.New(its.common.L(), strings.Join(paths, "/"))
		}
	}
	return node, nil
}

func (its *jsonPrimitive) getTargetFromPatch(path string) (jsonType, string, errors.OrdaError) {
	paths := strings.Split(path, "/")

	if len(paths) < 1 {
		return nil, "", errors.DatatypeInvalidPatch.New(its.common.L(), "incorrect path: %v", path)
	}
	key := paths[len(paths)-1]
	paths = paths[1 : len(paths)-1]

	target, err := its.getTargetByPaths(paths)
	if err != nil {
		return nil, "", err
	}
	return target, key, nil
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
		// Since a tombstone is placed in the Cemetery based on its.D, it should be adjusted.
		if tomb, ok := its.common.Cemetery[its.D.Hash()]; ok {
			delete(its.common.Cemetery, its.D.Hash())
			its.common.Cemetery[ts.Hash()] = tomb
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

func (its *jsonPrimitive) getLogger() *log.OrdaLog {
	return its.common.L()
}

func (its *jsonPrimitive) findJSONType(ts *model.Timestamp) (j jsonType, ok bool) {
	node, ok := its.common.NodeMap[ts.Hash()]
	return node, ok
}

func (its *jsonPrimitive) findJSONElement(ts *model.Timestamp) (j *jsonElement, ok bool) {
	if node, ok := its.common.NodeMap[ts.Hash()]; ok {
		if j, ok2 := node.(*jsonElement); ok2 {
			return j, ok2
		}
	}
	return nil, false
}

func (its *jsonPrimitive) findJSONObject(ts *model.Timestamp) (json *jsonObject, ok bool) {
	if node, ok := its.common.NodeMap[ts.Hash()]; ok {
		if j, ok2 := node.(*jsonObject); ok2 {
			return j, ok2
		}
	}
	return nil, false
}

func (its *jsonPrimitive) findJSONArray(ts *model.Timestamp) (json *jsonArray, ok bool) {
	if node, ok := its.common.NodeMap[ts.Hash()]; ok {
		if j, ok2 := node.(*jsonArray); ok2 {
			return j, ok2
		}
	}
	return nil, false
}

func (its *jsonPrimitive) addToNodeMap(primitive jsonType) {
	its.common.NodeMap[primitive.getCreateTime().Hash()] = primitive
}

func (its *jsonPrimitive) removeFromNodeMap(primitive jsonType) {
	delete(its.common.NodeMap, primitive.getCreateTime().Hash())
}

func (its *jsonPrimitive) getDeleteTime() *model.Timestamp {
	return its.D
}

func (its *jsonPrimitive) addToCemetery(primitive jsonType) {
	its.common.Cemetery[primitive.getDeleteTime().Hash()] = primitive
}

func (its *jsonPrimitive) getCommon() *jsonCommon {
	return its.common
}

func (its *jsonPrimitive) getRoot() *jsonObject {
	return its.common.root
}

func (its *jsonPrimitive) setRoot(r *jsonObject) {
	its.common.root = r
	its.common.NodeMap[r.getCreateTime().Hash()] = r
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
	kind := rv.Kind()
	switch kind {
	case reflect.Struct:
		toMap, err := utils.StructToMap(rv.Interface())
		if err != nil {
			its.common.L().Errorf("illegal struct: %v", err)
			return nil
		}
		return its.createJSONObject(parent, toMap, ts)
	case reflect.Map:
		return its.createJSONObject(parent, rv.Interface(), ts)
	case reflect.Slice, reflect.Array:
		return its.createJSONArray(parent, rv.Interface(), ts)
	case reflect.Ptr, reflect.Interface:
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
	ja := newJSONArray(its.common.BaseDatatype, parent, ts.GetAndNextDelimiter())
	var appendValues []timedType
	elements := reflect.ValueOf(value)
	for i := 0; i < elements.Len(); i++ {
		rv := elements.Index(i)
		if jt := its.createJSONTypeFromReflectValue(ja, rv, ts); jt != nil {
			appendValues = append(appendValues, jt)
		}
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

	jo := newJSONObject(its.common.BaseDatatype, parent, ts.GetAndNextDelimiter())
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
	if jt := its.createJSONTypeFromReflectValue(jo, value, ts); jt != nil {
		jo.putCommonWithTimedType(key, jt)
		its.addToNodeMap(jt)
	}
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
) (jsonType, errors.OrdaError) {
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
) (jsonType, errors.OrdaError) {
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
	errors.OrdaError, // error
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
) (jsonType, errors.OrdaError) {
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
) ([]*model.Timestamp, []jsonType, errors.OrdaError) {
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
) ([]jsonType, errors.OrdaError) {
	if parentArray, ok := its.findJSONArray(parent); ok {
		return parentArray.updateRemote(ts, targets, values)
	}
	return nil, errors.DatatypeInvalidParent.New(its.getLogger(), parent.ToString())
}

func (its *jsonPrimitive) DeleteLocalInArray(
	parent *model.Timestamp,
	pos, numOfNodes int,
	ts *model.Timestamp,
) ([]*model.Timestamp, []jsonType, errors.OrdaError) {
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
) ([]jsonType, errors.OrdaError) {
	if parentArray, ok := its.findJSONArray(parent); ok {
		return parentArray.deleteRemote(targets, ts)
	}
	return nil, errors.DatatypeInvalidParent.New(its.getLogger(), parent.ToString())
}

// ///////////////////// methods of iface.Snapshot ///////////////////////////////////////

func (its *jsonPrimitive) ToJSON() interface{} {
	return its.getValue()
}

func (its *jsonPrimitive) marshal() *marshaledJSONType {
	var p *model.Timestamp = nil
	if its.parent != nil {
		p = its.parent.getCreateTime()
	}
	return &marshaledJSONType{
		P: p,
		C: its.C,
		D: its.D,
	}
}

func (its *jsonPrimitive) unmarshal(marshaled *marshaledJSONType, assistant *unmarshalAssistant) {
	// do nothing
}
