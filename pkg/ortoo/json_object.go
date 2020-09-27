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
			T:      ts,
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
	removed, _ := its.putCommonWithTimedType(key, newChild) // by hashMapSnapshot

	if removed != nil {
		removedJSON := removed.(jsonType)
		/*
			The removedJSON.makeTomb(ts) should work as follows.
			JSONObject and JSONArray remain in NodeMap because they can be accessed as parents by other remote operations.
			Also, they should be added to Cemetery in order to be garbage-collected from NodeMap.
			JSONElement is immediately garbage-collected from NodeMap because it is never accessed.
			Hence, it is not added to Cemetery.

			Even if any jsonType is already a tombstone, it is not deleted again.

			jsonElement: removed from NodeMap, not added to Cemetery.
			jsonObject: remain in NodeMap, added to Cemetery.
			jsonArray: remain in NodeMap, added to Cemetery.
		*/
		if !removedJSON.isTomb() {
			removedJSON.makeTomb(ts)
		}
		if je, ok := removed.(*jsonElement); ok {
			its.removeFromNodeMap(je)
		} else {
			its.addToCemetery(removedJSON)
		}
		return removedJSON
	}
	return nil
}

func (its *jsonObject) DeleteLocalInObject(
	parent *model.Timestamp,
	key string,
	ts *model.Timestamp,
) (jsonType, errors.OrtooError) {
	return its.deleteInObject(parent, key, ts, true)
}

func (its *jsonObject) DeleteRemoteInObject(
	parent *model.Timestamp,
	key string,
	ts *model.Timestamp,
) (jsonType, errors.OrtooError) {
	return its.deleteInObject(parent, key, ts, false)
}

func (its *jsonObject) deleteInObject(
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
	[]interface{}, // inserted values
	errors.OrtooError, // error
) {
	if parentArray, ok := its.findJSONArray(parent); ok {
		return parentArray.insertCommon(pos, nil, ts, values...)
	}
	return nil, nil, errors.ErrDatatypeInvalidParent.New(its.getLogger(), parent.ToString())
}

func (its *jsonObject) InsertRemoteInArray(
	parent *model.Timestamp,
	target *model.Timestamp,
	ts *model.Timestamp,
	values ...interface{},
) errors.OrtooError {
	if parentArray, ok := its.findJSONArray(parent); ok {
		_, _, err := parentArray.insertCommon(-1, target, ts, values...)
		return err
	}
	return errors.ErrDatatypeInvalidParent.New(its.getLogger(), parent.ToString())
}

func (its *jsonObject) UpdateLocalInArray(
	parent *model.Timestamp,
	pos int,
	ts *model.Timestamp,
	values ...interface{},
) ([]*model.Timestamp, errors.OrtooError) {
	if parentArray, ok := its.findJSONArray(parent); ok {
		return parentArray.updateLocal(pos, ts, values...)
	}
	return nil, errors.ErrDatatypeInvalidParent.New(its.getLogger(), parent.ToString())
}

func (its *jsonObject) UpdateRemoteInArray(
	parent *model.Timestamp,
	ts *model.Timestamp,
	targets []*model.Timestamp,
	values []interface{},
) []errors.OrtooError {
	if parentArray, ok := its.findJSONArray(parent); ok {
		return parentArray.updateRemote(ts, targets, values)
	}
	return []errors.OrtooError{errors.ErrDatatypeInvalidParent.New(its.getLogger(), parent.ToString())}
}

func (its *jsonObject) DeleteLocalInArray(
	parent *model.Timestamp,
	pos, numOfNodes int,
	ts *model.Timestamp,
) ([]*model.Timestamp, []interface{}, errors.OrtooError) {
	if parentArray, ok := its.findJSONArray(parent); ok {
		return parentArray.deleteLocal(pos, numOfNodes, ts)
	}
	return nil, nil, errors.ErrDatatypeInvalidParent.New(its.getLogger(), parent.ToString())
}

func (its *jsonObject) DeleteRemoteInArray(
	parent *model.Timestamp,
	ts *model.Timestamp,
	targets []*model.Timestamp,
) []errors.OrtooError {
	if parentArray, ok := its.findJSONArray(parent); ok {
		return parentArray.deleteRemote(targets, ts)
	}
	return []errors.OrtooError{errors.ErrDatatypeInvalidParent.New(its.getLogger(), parent.ToString())}
}

func (its *jsonObject) getValue() types.JSONValue {
	return its.GetAsJSONCompatible()
}

func (its *jsonObject) getType() TypeOfJSON {
	return TypeJSONObject
}

// func (its *jsonObject) makeTombAsChild(ts *model.Timestamp) bool {
// 	if its.jsonType.makeTombAsChild(ts) {
// 		its.addToCemetery(its)
// 		for _, v := range its.Map {
// 			cast := v.(jsonType)
// 			cast.makeTombAsChild(ts)
// 		}
// 		return true
// 	}
// 	return false
// }

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
		parentTS = parent.getKeyTime().ToString()
	}
	return fmt.Sprintf("JO(%v)[T%v|V%v]", parentTS, its.getKeyTime().ToString(), its.hashMapSnapshot.String())
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
