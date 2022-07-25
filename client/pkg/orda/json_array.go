package orda

import (
	"fmt"
	errors2 "github.com/orda-io/orda/client/pkg/errors"
	"github.com/orda-io/orda/client/pkg/iface"
	model2 "github.com/orda-io/orda/client/pkg/model"
	"github.com/orda-io/orda/client/pkg/types"
)

// ////////////////////////////////////
//  jsonArray
// ////////////////////////////////////

type jsonArray struct {
	jsonType
	*listSnapshot
}

func newJSONArray(base iface.BaseDatatype, parent jsonType, ts *model2.Timestamp) *jsonArray {
	return &jsonArray{
		jsonType: &jsonPrimitive{
			parent: parent,
			common: parent.getCommon(),
			C:      ts,
		},
		listSnapshot: newListSnapshot(base),
	}
}

func (its *jsonArray) getManyJSONTypes(pos int, numOfNodes int) []jsonType {
	tts := its.findManyTimedTypes(pos, numOfNodes)
	var ret []jsonType
	for _, tt := range tts {
		ret = append(ret, tt.(jsonType))
	}
	return ret
}

func (its *jsonArray) getJSONType(pos int) jsonType {
	tt := its.findTimedType(pos)
	return tt.(jsonType)
}

func (its *jsonArray) insertCommon(
	pos int, // in the case of the local insert
	target *model2.Timestamp, // in the case of the remote insert
	ts *model2.Timestamp,
	values ...interface{},
) (*model2.Timestamp, []interface{}, errors2.OrdaError) {
	var inserted []timedType
	for _, v := range values {
		ins := its.createJSONType(its, v, ts)
		inserted = append(inserted, ins)
		its.addToNodeMap(ins)
	}
	if target == nil {
		t, ins := its.listSnapshot.insertLocalWithTimedTypes(pos, inserted...)
		return t, ins, nil
	}
	return nil, nil, its.listSnapshot.insertRemoteWithTimedTypes(target, inserted...)
}

func (its *jsonArray) deleteLocal(
	pos int,
	numOfNodes int,
	ts *model2.Timestamp,
) ([]*model2.Timestamp, []jsonType) {
	targets, timedTypes, _ := its.listSnapshot.deleteLocal(pos, numOfNodes, ts)
	for _, v := range targets {
		if jt, ok := its.findJSONType(v); ok {
			its.addToCemetery(jt)
		}
	}
	var jsonTypes []jsonType
	for _, t := range timedTypes {
		jsonTypes = append(jsonTypes, t.(jsonType))
	}
	return targets, jsonTypes
}

func (its *jsonArray) deleteRemote(
	targets []*model2.Timestamp,
	ts *model2.Timestamp,
) (deleted []jsonType, err errors2.OrdaError) {
	deletedTimedTypes, err := its.listSnapshot.deleteRemote(targets, ts)
	for _, t := range deletedTimedTypes {
		jt := t.(jsonType)
		deleted = append(deleted, jt)
		its.addToCemetery(jt)
	}
	return
}

func (its *jsonArray) updateLocal(
	pos int,
	ts *model2.Timestamp,
	values ...interface{},
) ([]*model2.Timestamp, []jsonType, errors2.OrdaError) {
	var updatedTargets []*model2.Timestamp
	var oldJSONTypes []jsonType
	target := its.retrieve(pos + 1)
	for _, v := range values {
		/*
			In the list, orderedType.K is used to resolve the order conflicts, and to find targets by remote operations.
			The new node should preserve orderedType.K, and thus its timedType is replaced with the new jsonType.
			In addition, the old jsonType except jsonElement should be accessible by some remote operations as a parent.
			Thus, they are added to Cemetery.
		*/
		oldOne := target.getTimedType().(jsonType)
		updatedTargets = append(updatedTargets, target.getOrderTime())
		newOne := its.createJSONType(its, v, ts) // ts's delimiter might increase.
		target.setTimedType(newOne)
		its.addToNodeMap(newOne)
		its.funeral(oldOne, newOne.getTime())
		oldJSONTypes = append(oldJSONTypes, oldOne)
		target = target.getNextLive()
	}
	return updatedTargets, oldJSONTypes, nil
}

func (its *jsonArray) updateRemote(
	ts *model2.Timestamp,
	targets []*model2.Timestamp,
	values []interface{},
) ([]jsonType, errors2.OrdaError) {
	errs := &errors2.MultipleOrdaErrors{}
	var delTypes []jsonType = nil
	for i, t := range targets {
		newOne := its.createJSONType(its, values[i], ts)
		its.addToNodeMap(newOne)
		// thisTS := ts.GetAndNextDelimiter()
		if node, ok := its.Map[t.Hash()]; ok {
			var deleted, updated jsonType
			oldOne := node.getTimedType().(jsonType)
			if !node.isTomb() {
				if node.getTime().Compare(newOne.getCreateTime()) < 0 {
					node.setTimedType(newOne)
					deleted = oldOne
					updated = newOne
				} else {
					deleted = newOne
					updated = oldOne
				}
			} else { // tombstone is not recovered.
				deleted = newOne
				updated = oldOne
			}
			its.funeral(deleted, updated.getCreateTime())
			delTypes = append(delTypes, deleted)
		} else {
			_ = errs.Append(errors2.DatatypeNoTarget.New(its.L(), t.ToString()))
		}
	}
	return delTypes, errs.Return()
}

func (its *jsonArray) getValue() types.JSONValue {
	return its.ToJSON()
}

func (its *jsonArray) getType() TypeOfJSON {
	return TypeJSONArray
}

func (its *jsonArray) setValue(v types.JSONValue) {
	panic("not used")
}

func (its *jsonArray) String() string {
	parent := its.getParent()
	parentTS := "nil"
	if parent != nil {
		parentTS = parent.getCreateTime().ToString()
	}
	return fmt.Sprintf("JA(%v)[C%v|V%v", parentTS, its.getCreateTime().ToString(), its.listSnapshot.String())
}

func (its *jsonArray) equal(o jsonType) bool {
	if its.getType() != o.getType() {
		return false
	}

	ja := o.(*jsonArray)
	if !its.jsonType.equal(ja.jsonType) {
		return false
	}
	if its.size != ja.size {
		return false
	}
	for k, v1 := range its.Map {
		v2 := ja.Map[k]
		if (v1 == nil && v2 != nil) || (v1 != nil && v2 == nil) {
			return false
		}
		if v1 == nil && v2 == nil { // cannot happen
			return false
		}
		ov1, ov2 := v1.(orderedType), v2.(orderedType)
		if ov1.getOrderTime().Compare(ov2.getOrderTime()) != 0 {
			return false
		}
		if ov1.getOrderTime().Compare(model2.OldestTimestamp()) == 0 &&
			ov1.getOrderTime().Compare(ov2.getOrderTime()) == 0 {
			return true
		}
		jt1, jt2 := ov1.getTimedType().(jsonType), ov2.getTimedType().(jsonType)
		if jt1.getCreateTime().Compare(jt2.getCreateTime()) != 0 {
			return false
		}
		if (ov1.getPrev() == nil && ov2.getPrev() != nil) || (ov1.getPrev() != nil && ov2.getPrev() == nil) {
			return false
		}
		if ov1.getPrev() != nil && ov2.getPrev() != nil {
			if ov1.getPrev().getOrderTime().Compare(ov2.getPrev().getOrderTime()) != 0 {
				return false
			}
		}
		if (ov1.getNext() == nil && ov2.getNext() != nil) || (ov1.getNext() != nil && ov2.getNext() == nil) {
			return false
		}
		if ov1.getNext() != nil && ov2.getNext() != nil {
			if ov1.getNext().getOrderTime().Compare(ov2.getNext().getOrderTime()) != 0 {
				return false
			}
		}
	}
	return true
}

// ///////////////////// methods of iface.Snapshot ///////////////////////////////////////
//
// func (its *jsonArray) GetBase() iface.BaseDatatype {
// 	return its.getCommon().base
// }
//
// func (its *jsonArray) SetBase(base iface.BaseDatatype) {
// 	its.getCommon().SetBase(base)
// }
//
// func (its *jsonArray) CloneSnapshot() iface.Snapshot {
// 	panic("Implement me")
// }

// ToJSON returns an interface type that contains all live objects.
func (its *jsonArray) ToJSON() interface{} {
	var list = make([]interface{}, 0)
	n := its.listSnapshot.head.getNextLive()
	for n != nil {
		if !n.isTomb() {
			switch cast := n.getTimedType().(type) {
			case *jsonObject:
				list = append(list, cast.ToJSON())
			case *jsonElement:
				list = append(list, cast.getValue())
			case *jsonArray:
				list = append(list, cast.ToJSON())
			}
		}
		n = n.getNextLive()
	}
	return list
}

func (its *jsonArray) marshal() *marshaledJSONType {
	marshal := its.jsonType.marshal()
	marshal.T = marshalKeyJSONArray
	marshaledJA := &marshaledJSONArray{
		S: its.listSnapshot.size,
	}
	n := its.listSnapshot.head.getNext() // NOT store head
	for n != nil {
		jt := n.getTimedType().(jsonType)
		var mot marshaledOrderedType
		if n.getOrderTime() == jt.getCreateTime() {
			mot = [2]*model2.Timestamp{n.getOrderTime()}
		} else {
			mot = [2]*model2.Timestamp{n.getOrderTime(), jt.getCreateTime()}
		}

		marshaledJA.N = append(marshaledJA.N, mot)
		n = n.getNext()
	}
	marshal.A = marshaledJA
	return marshal
}

func (its *jsonArray) unmarshal(marshaled *marshaledJSONType, assistant *unmarshalAssistant) {
	marshaledJA := marshaled.A
	its.listSnapshot = newListSnapshot(assistant.common.BaseDatatype)
	prev := its.listSnapshot.head
	for _, mot := range marshaledJA.N {
		o := mot[0]
		c := mot[1]
		if c == nil {
			c = o
		}
		timedType, _ := its.findJSONType(c)
		node := &orderedNode{
			timedType: timedType,
			O:         assistant.unifyTimestamp(o),
		}
		its.Map[node.getOrderTime().Hash()] = node
		prev.insertNext(node)
		prev = node

	}
	its.size = marshaled.A.S
}
