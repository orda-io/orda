package ortoo

import (
	"fmt"
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/iface"
	"github.com/knowhunger/ortoo/pkg/internal/datatypes"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/pkg/types"
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
			C:      ts,
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

func (its *jsonObject) PutCommonInObject(
	parent *model.Timestamp,
	key string,
	value interface{},
	ts *model.Timestamp,
) (jsonType, errors.OrtooError) {
	if parentObj, ok := its.findJSONObject(parent); ok {
		return parentObj.putCommon(key, value, ts), nil
	}
	return nil, errors.ErrDatatypeInvalidParent.New(its.getLogger(), parent.ToString())
}

func (its *jsonObject) putCommon(key string, value interface{}, ts *model.Timestamp) jsonType {
	newChild := its.createJSONType(its, value, ts)
	its.addToNodeMap(newChild)
	// removed can be either the existing one or newChild.
	removed, put := its.putCommonWithTimedType(key, newChild) // by hashMapSnapshot

	if removed != nil {
		removedJSON := removed.(jsonType)
		putJSON := put.(jsonType)
		/*
			The removedJSON.makeTomb(ts) should work as follows.
			JSONObject and JSONArray remain in NodeMap because they can be accessed as parents by other remote operations.
			Also, they should be added to Cemetery in order to be garbage-collected from NodeMap.
			JSONElement is immediately garbage-collected from NodeMap because it is never accessed.
			Hence, it is not added to Cemetery.

			Even if any jsonType is already a tombstone, it is not deleted again.

			jsonElement: removed from NodeMap, not added to Cemetery.
			jsonObject, jsonArray: remain in NodeMap, added to Cemetery.
		*/
		its.funeral(removedJSON, putJSON.getCreateTime())
		return removedJSON
	}
	return nil
}

func (its *jsonObject) DeleteLocalInObject(
	parent *model.Timestamp,
	key string,
	ts *model.Timestamp,
) (jsonType, errors.OrtooError) {
	return its.deleteCommonInObject(parent, key, ts, true)
}

func (its *jsonObject) DeleteRemoteInObject(
	parent *model.Timestamp,
	key string,
	ts *model.Timestamp,
) (jsonType, errors.OrtooError) {
	return its.deleteCommonInObject(parent, key, ts, false)
}

func (its *jsonObject) deleteCommonInObject(
	parent *model.Timestamp,
	key string,
	ts *model.Timestamp,
	isLocal bool,
) (jsonType, errors.OrtooError) {
	if parentObj, ok := its.findJSONObject(parent); ok {
		/*
			If jsonType is deleted in jsonObject, it should remain in NodeMap, and be added to Cemetery.
		*/
		var deleted timedType
		var err errors.OrtooError
		if isLocal {
			deleted, _, err = parentObj.removeLocalWithTimedType(key, ts)
		} else {
			deleted, _, err = parentObj.removeRemoteWithTimedType(key, ts)
		}
		if deleted != nil {
			deletedJT := deleted.(jsonType)
			its.addToCemetery(deletedJT)
			return deletedJT, nil
		}
		return nil, err
	}
	return nil, errors.ErrDatatypeInvalidParent.New(its.getLogger(), parent.ToString())
}

func (its *jsonObject) getAsJSONType(key string) jsonType {
	if v, ok := its.Map[key]; ok {
		return v.(jsonType)
	}
	return nil
}

// InsertLocalInArray inserts values locally and returns the timestamp of target
func (its *jsonObject) InsertLocalInArray(
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
	return nil, nil, errors.ErrDatatypeInvalidParent.New(its.getLogger(), parent.ToString())
}

func (its *jsonObject) InsertRemoteInArray(
	parent *model.Timestamp,
	target *model.Timestamp,
	ts *model.Timestamp,
	values ...interface{},
) (jsonType, errors.OrtooError) {
	if parentArray, ok := its.findJSONArray(parent); ok {
		_, _, err := parentArray.insertCommon(-1, target, ts, values...)
		return parentArray, err
	}
	return nil, errors.ErrDatatypeInvalidParent.New(its.getLogger(), parent.ToString())
}

func (its *jsonObject) UpdateLocalInArray(
	parent *model.Timestamp,
	pos int,
	ts *model.Timestamp,
	values ...interface{},
) ([]*model.Timestamp, []jsonType, errors.OrtooError) {
	if parentArray, ok := its.findJSONArray(parent); ok {
		return parentArray.updateLocal(pos, ts, values...)
	}
	return nil, nil, errors.ErrDatatypeInvalidParent.New(its.getLogger(), parent.ToString())
}

func (its *jsonObject) UpdateRemoteInArray(
	parent *model.Timestamp,
	ts *model.Timestamp,
	targets []*model.Timestamp,
	values []interface{},
) ([]jsonType, errors.OrtooError) {
	if parentArray, ok := its.findJSONArray(parent); ok {
		return parentArray.updateRemote(ts, targets, values)
	}
	return nil, errors.ErrDatatypeInvalidParent.New(its.getLogger(), parent.ToString())
}

func (its *jsonObject) DeleteLocalInArray(
	parent *model.Timestamp,
	pos, numOfNodes int,
	ts *model.Timestamp,
) ([]*model.Timestamp, []jsonType, errors.OrtooError) {
	if parentArray, ok := its.findJSONArray(parent); ok {
		return parentArray.deleteLocal(pos, numOfNodes, ts)
	}
	return nil, nil, errors.ErrDatatypeInvalidParent.New(its.getLogger(), parent.ToString())
}

func (its *jsonObject) DeleteRemoteInArray(
	parent *model.Timestamp,
	ts *model.Timestamp,
	targets []*model.Timestamp,
) ([]jsonType, errors.OrtooError) {
	if parentArray, ok := its.findJSONArray(parent); ok {
		return parentArray.deleteRemote(targets, ts)
	}
	return nil, errors.ErrDatatypeInvalidParent.New(its.getLogger(), parent.ToString())
}

func (its *jsonObject) getValue() types.JSONValue {
	return its.GetAsJSONCompatible()
}

func (its *jsonObject) getType() TypeOfJSON {
	return TypeJSONObject
}

func (its *jsonObject) getChildAsJSONElement(key string) *jsonElement {
	value := its.getFromMap(key)
	if value == nil {
		return nil
	}
	return value.(*jsonElement)
}

func (its *jsonObject) getChildAsJSONObject(key string) *jsonObject {
	value := its.getFromMap(key)
	return value.(*jsonObject)
}

func (its *jsonObject) getChildAsJSONArray(key string) *jsonArray {
	value := its.getFromMap(key)
	return value.(*jsonArray)
}

func (its *jsonObject) String() string {
	parent := its.getParent()
	parentTS := "nil"
	if parent != nil {
		parentTS = parent.getCreateTime().ToString()
	}
	return fmt.Sprintf("JO(%v)[C%v|V%v]", parentTS, its.getCreateTime().ToString(), its.hashMapSnapshot.String())
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

func (its *jsonObject) Equal(o *jsonObject) bool {
	return its.jsonType.(*jsonPrimitive).common.equal(o.jsonType.(*jsonPrimitive).common)
}

func (its *jsonObject) equal(o jsonType) bool {
	if its.getType() != o.getType() {
		return false
	}
	jo := o.(*jsonObject)
	if !its.jsonType.equal(jo.jsonType) {
		return false
	}

	if its.Size != jo.Size {
		return false
	}
	for k, v1 := range its.Map {
		v2 := jo.Map[k]
		if (v1 == nil && v2 != nil) || (v1 != nil && v2 == nil) {
			return false
		}
		if v1 == nil && v2 == nil {
			continue
		}
		jv1, jv2 := v1.(jsonType), v2.(jsonType)
		if jv1.getCreateTime().Compare(jv2.getCreateTime()) != 0 {
			return false
		}
	}
	return true
}
