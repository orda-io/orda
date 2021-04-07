package ortoo

import (
	"fmt"
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/iface"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/pkg/types"
)

// ////////////////////////////////////
//  jsonObject
// ////////////////////////////////////

type jsonObject struct {
	jsonType
	*mapSnapshot
}

func newJSONObject(base iface.BaseDatatype, parent jsonType, ts *model.Timestamp) *jsonObject {
	var root *jsonCommon
	if parent == nil {
		root = &jsonCommon{
			root:     nil,
			base:     base,
			nodeMap:  make(map[string]jsonType),
			cemetery: make(map[string]jsonType),
		}
	} else {
		root = parent.getCommon()
	}
	obj := &jsonObject{
		jsonType: &jsonPrimitive{
			common: root,
			parent: parent,
			C:      ts,
		},
		mapSnapshot: newMapSnapshot(base),
	}
	if parent == nil {
		obj.jsonType.setRoot(obj)
	}
	return obj
}

func (its *jsonObject) putCommon(key string, value interface{}, ts *model.Timestamp) jsonType {
	newChild := its.createJSONType(its, value, ts)
	its.addToNodeMap(newChild)
	// removed can be either the existing one or newChild.
	removed, put := its.putCommonWithTimedType(key, newChild) // by mapSnapshot

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

func (its *jsonObject) deleteCommonInObject(
	key string,
	ts *model.Timestamp,
	isLocal bool,
) (jsonType, errors.OrtooError) {
	/*
		If jsonType is deleted in jsonObject, it should remain in NodeMap, and be added to Cemetery.
	*/
	var deleted timedType
	var err errors.OrtooError
	if isLocal {
		deleted, _, err = its.removeLocalWithTimedType(key, ts)
	} else {
		deleted, _, err = its.removeRemoteWithTimedType(key, ts)
	}
	if deleted != nil {
		deletedJT := deleted.(jsonType)
		its.addToCemetery(deletedJT)
		return deletedJT, nil
	}
	return nil, err
}

func (its *jsonObject) getAsJSONType(key string) jsonType {
	if v, ok := its.Map[key]; ok {
		return v.(jsonType)
	}
	return nil
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
	return fmt.Sprintf("JO(%v)[C%v|V%v]", parentTS, its.getCreateTime().ToString(), its.mapSnapshot.String())
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

// ///////////////////// methods of iface.Snapshot ///////////////////////////////////////

func (its *jsonObject) GetBase() iface.BaseDatatype {
	return its.getCommon().base
}

func (its *jsonObject) SetBase(base iface.BaseDatatype) {
	its.getCommon().SetBase(base)
}

func (its *jsonObject) CloneSnapshot() iface.Snapshot {
	// TODO: implement CloneSnapshot()
	return &jsonObject{}
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
