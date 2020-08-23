package ortoo

import (
	"fmt"
	"github.com/knowhunger/ortoo/ortoo/errors"
	"github.com/knowhunger/ortoo/ortoo/iface"
	"github.com/knowhunger/ortoo/ortoo/internal/datatypes"
	"github.com/knowhunger/ortoo/ortoo/model"
	"github.com/knowhunger/ortoo/ortoo/operations"
	"github.com/knowhunger/ortoo/ortoo/types"
	"reflect"
)

// ////////////////////////////////////
//  jsonObject
// ////////////////////////////////////

type jsonObject struct {
	jsonType
	*hashMapSnapshot
}

func newJSONObject(base *datatypes.BaseDatatype, parent jsonType, ts *model.Timestamp) *jsonObject {
	var root *jsonCommon
	if parent == nil {
		root = &jsonCommon{
			root:     nil,
			base:     base,
			nodeMap:  make(map[string]jsonType),
			cemetery: make(map[string]jsonType),
		}
	} else {
		root = parent.getRoot()
	}
	obj := &jsonObject{
		jsonType: &jsonPrimitive{
			common: root,
			parent: parent,
			K:      ts,
			P:      ts,
		},
		hashMapSnapshot: newHashMapSnapshot(base),
	}
	if parent == nil {
		obj.jsonType.setRoot(obj)
	}
	return obj
}

func (its *jsonObject) CloneSnapshot() iface.Snapshot {
	// TODO: implement CloneSnapshot()
	return &jsonObject{}
}

func (its *jsonObject) PutCommonInObject(parent *model.Timestamp, key string, value interface{}, ts *model.Timestamp) (jsonType, errors.OrtooError) {
	if parentObj, ok := its.findJSONObject(parent); ok {
		return parentObj.putCommon(key, value, ts), nil
	}
	return nil, errors.ErrDatatypeInvalidParent.New(its.getLogger())
}

func (its *jsonObject) putCommon(key string, value interface{}, ts *model.Timestamp) jsonType {
	rt := reflect.ValueOf(value)
	var newChild jsonType
	switch rt.Kind() {
	case reflect.Slice, reflect.Array:
		newChild = its.createJSONArray(its, value, ts)
	case reflect.Struct, reflect.Map:
		newChild = its.createJSONObject(its, value, ts)
	case reflect.Ptr:
		val := rt.Elem()
		newChild = its.putCommon(key, val.Interface(), ts) // recursively
	default:
		newChild = newJSONElement(its, types.ConvertToJSONSupportedValue(value), ts.NextDeliminator())
	}
	removed, _ := its.putCommonWithTimedValue(key, newChild) // in hash map
	if removed != nil {
		its.addToCemetery(removed.(jsonType))
	}
	its.addToNodeMap(newChild)
	return newChild
}

func (its *jsonObject) DeleteLocalInObject(parent *model.Timestamp, key string, ts *model.Timestamp) (interface{}, errors.OrtooError) {
	if parentObj, ok := its.findJSONObject(parent); ok {
		return parentObj.removeLocal(key, ts)
	}
	return nil, errors.ErrDatatypeInvalidParent.New(its.getLogger())
}

func (its *jsonObject) DeleteRemoteInObject(parent *model.Timestamp, key string, ts *model.Timestamp) (interface{}, errors.OrtooError) {
	if parentObj, ok := its.findJSONObject(parent); ok {
		ret := parentObj.removeRemote(key, ts)
		return ret, nil
	}
	return nil, errors.ErrDatatypeInvalidParent.New(its.getLogger())
}

func (its *jsonObject) getAsJSONType(key string) jsonType {
	if v, ok := its.Map[key]; ok {
		return v.(jsonType)
	}
	return nil
}

// InsertLocal inserts values locally and returns the timestamp of target
func (its *jsonObject) InsertLocal(
	parent *model.Timestamp,
	pos int,
	ts *model.Timestamp,
	values ...interface{},
) (
	*model.Timestamp, // the timestamp of target
	[]interface{}, // inserted values
	errors.OrtooError, // error
) {
	if parentArray, ok := its.findJSONArray(parent); ok {
		return parentArray.arrayInsertCommon(pos, nil, ts, values...)
	}
	return nil, nil, errors.ErrDatatypeInvalidParent.New(its.getLogger())
}

func (its *jsonObject) InsertRemote(parent *model.Timestamp, target, ts *model.Timestamp, values ...interface{}) {
	if parentArray, ok := its.findJSONArray(parent); ok {
		_, _, _ = parentArray.arrayInsertCommon(-1, target, ts, values...)
		return
	}
	_ = errors.ErrDatatypeInvalidParent.New(its.getLogger())
}

func (its *jsonObject) UpdateLocalInArray(op *operations.DocUpdateInArrayOperation) {
	// if parentArray, ok := its.findJSONArray(op.C.P); ok {
	// 	return parentArray.arrayUpdateLocal(op)
	// }
	// return nil, nil, errors.New(errors.ErrDatatypeInvalidParent)
}

func (its *jsonObject) DeleteLocalInArray(
	parent *model.Timestamp, pos, numOfNodes int, ts *model.Timestamp) ([]*model.Timestamp, []interface{}, errors.OrtooError) {
	if parentArray, ok := its.findJSONArray(parent); ok {
		return parentArray.arrayDeleteLocal(pos, numOfNodes, ts)
	}
	return nil, nil, errors.ErrDatatypeInvalidParent.New(its.getLogger())
}

func (its *jsonObject) DeleteRemoteInArray(parent *model.Timestamp, targets []*model.Timestamp, ts *model.Timestamp) {
	if parentArray, ok := its.findJSONArray(parent); ok {
		parentArray.arrayDeleteRemote(targets, ts)
		return
	}
	_ = errors.ErrDatatypeInvalidParent.New(its.getLogger())
}

func (its *jsonObject) getValue() types.JSONValue {
	return its.GetAsJSONCompatible()
}

func (its *jsonObject) getType() TypeOfJSON {
	return TypeJSONObject
}

func (its *jsonObject) makeTombAsChild(ts *model.Timestamp) bool {
	if its.jsonType.makeTombAsChild(ts) {
		its.addToCemetery(its)
		for _, v := range its.Map {
			cast := v.(jsonType)
			cast.makeTombAsChild(ts)
		}
		return true
	}
	return false
}

func (its *jsonObject) makeTomb(ts *model.Timestamp) bool {
	if its.jsonType.makeTomb(ts) {
		its.addToCemetery(its)
		for _, v := range its.Map {
			cast := v.(jsonType)
			cast.makeTombAsChild(ts)
		}
		return true
	}
	return false
}

func (its *jsonObject) getChildAsJSONElement(key string) *jsonElement {
	value := its.get(key)
	if value == nil {
		return nil
	}
	return value.(*jsonElement)
}

func (its *jsonObject) getChildAsJSONObject(key string) *jsonObject {
	value := its.get(key)
	return value.(*jsonObject)
}

func (its *jsonObject) getChildAsJSONArray(key string) *jsonArray {
	value := its.get(key)
	return value.(*jsonArray)
}

func (its *jsonObject) String() string {
	parent := its.getParent()
	parentTS := "nil"
	if parent != nil {
		parentTS = parent.getKey().ToString()
	}
	return fmt.Sprintf("JO(%v)[T%v|V%v]", parentTS, its.getKey().ToString(), its.hashMapSnapshot.String())
}

func (its *jsonObject) GetAsJSONCompatible() interface{} {
	m := make(map[string]interface{})
	for k, v := range its.Map {
		if v != nil {
			switch cast := v.(type) {
			case *jsonObject:
				if !cast.isTomb() {
					m[k] = cast.GetAsJSONCompatible()
				}

			case *jsonElement:
				if !cast.isTomb() {
					m[k] = v.getValue()
				}
			case *jsonArray:
				if !cast.isTomb() {
					m[k] = cast.GetAsJSONCompatible()
				}
			}
		}
	}
	return m
}
