package orda

import (
	"encoding/json"
	"fmt"

	"github.com/orda-io/orda/pkg/errors"
	"github.com/orda-io/orda/pkg/iface"
	"github.com/orda-io/orda/pkg/model"
	"github.com/orda-io/orda/pkg/types"
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
			BaseDatatype: base,
			root:         nil,
			NodeMap:      make(map[string]jsonType),
			Cemetery:     make(map[string]jsonType),
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
) (jsonType, errors.OrdaError) {
	/*
		If jsonType is deleted in jsonObject, it should remain in NodeMap, and be added to Cemetery.
	*/
	var deleted timedType
	var err errors.OrdaError
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
	return its.ToJSON()
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
		if !jv1.equal(jv2) {
			return false
		}
	}
	return true
}

// ///////////////////// methods of iface.Snapshot ///////////////////////////////////////

// MarshalJSON returns marshaledDocument.
func (its *jsonObject) MarshalJSON() ([]byte, error) {
	marshalDoc := newMarshaledDocument()
	for _, v := range its.getCommon().NodeMap {
		marshaled := v.marshal()
		marshalDoc.NodeMap = append(marshalDoc.NodeMap, marshaled)
	}
	return json.Marshal(marshalDoc)
}

func (its *jsonObject) UnmarshalJSON(bytes []byte) error {
	var forUnmarshal marshaledDocument

	if err := json.Unmarshal(bytes, &forUnmarshal); err != nil {
		return errors.DatatypeMarshal.New(its.getLogger(), err.Error())
	}

	assistant := &unmarshalAssistant{
		tsMap: make(map[string]*model.Timestamp),
		common: &jsonCommon{
			BaseDatatype: its.BaseDatatype,
			root:         its,
			NodeMap:      make(map[string]jsonType),
			Cemetery:     make(map[string]jsonType),
		},
	}
	// make all skeleton jsonTypes "in advance"
	oldestTs := model.OldestTimestamp()
	for _, v := range forUnmarshal.NodeMap {
		jt := v.unmarshalAsJSONType(assistant)
		if jt.getCreateTime().Compare(oldestTs) == 0 {
			its.jsonType = &jsonPrimitive{
				common: assistant.common,
				parent: nil,
				C:      jt.getCreateTime(),
				D:      jt.getDeleteTime(),
			}
			its.addToNodeMap(its)
		} else {
			jt.addToNodeMap(jt)
		}
	}

	// fill up the missing values for each jsonType
	for _, marshaled := range forUnmarshal.NodeMap {
		jt, _ := its.findJSONType(marshaled.C) // real jsonType
		if marshaled.P != nil {
			parent, _ := its.findJSONType(marshaled.P)
			jt.setParent(parent)
		}
		jt.unmarshal(marshaled, assistant) // unmarshal type-dependently
		if jt.isTomb() {
			its.addToCemetery(jt)
		}
	}
	return nil
}

func (its *jsonObject) marshal() *marshaledJSONType {
	marshal := its.jsonType.marshal()
	marshal.T = marshalKeyJSONObject
	value := &marshaledJSONObject{
		S: its.Size,
		M: make(map[string]*model.Timestamp),
	}
	for k, v := range its.mapSnapshot.Map {
		jt := v.(jsonType)
		value.M[k] = jt.getCreateTime() // store only create timestamp
	}
	marshal.O = value
	return marshal
}

func (its *jsonObject) unmarshal(marshaled *marshaledJSONType, assistant *unmarshalAssistant) {
	marshaledJO := marshaled.O
	its.mapSnapshot = &mapSnapshot{
		BaseDatatype: assistant.common.BaseDatatype,
		Map:          make(map[string]timedType),
		Size:         marshaledJO.S,
	}
	for k, ts := range marshaledJO.M {
		realInstance, _ := its.findJSONType(ts)
		its.mapSnapshot.Map[k] = realInstance
	}
}

func (its *jsonObject) ToJSON() interface{} {
	m := make(map[string]interface{})
	for k, v := range its.Map {
		if v != nil {
			if !v.isTomb() {
				switch cast := v.(type) {
				case *jsonObject:
					m[k] = cast.ToJSON()
				case *jsonElement:
					m[k] = v.getValue()
				case *jsonArray:
					m[k] = cast.ToJSON()
				}
			}
		}
	}
	return m
}
