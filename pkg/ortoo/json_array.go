package ortoo

import (
	"fmt"
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/iface"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/pkg/types"
)

// ////////////////////////////////////
//  jsonArray
// ////////////////////////////////////

type jsonArray struct {
	jsonType
	*listSnapshot
}

func newJSONArray(base iface.BaseDatatype, parent jsonType, ts *model.Timestamp) *jsonArray {
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
	target *model.Timestamp, // in the case of the remote insert
	ts *model.Timestamp,
	values ...interface{},
) (*model.Timestamp, []interface{}, errors.OrtooError) {
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
	return nil, nil, its.listSnapshot.insertRemoteWithTimedTypes(target, ts, inserted...)
}

func (its *jsonArray) deleteLocal(
	pos int,
	numOfNodes int,
	ts *model.Timestamp,
) ([]*model.Timestamp, []jsonType) {
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
	targets []*model.Timestamp,
	ts *model.Timestamp,
) (deleted []jsonType, err errors.OrtooError) {
	deletedTimedTypes, err := its.listSnapshot.deleteRemote(targets, ts)
	if err != nil {
		return nil, err
	}
	for _, t := range deletedTimedTypes {
		jt := t.(jsonType)
		deleted = append(deleted, jt)
		its.addToCemetery(jt)
	}
	return
}

func (its *jsonArray) updateLocal(
	pos int,
	ts *model.Timestamp,
	values ...interface{},
) ([]*model.Timestamp, []jsonType, errors.OrtooError) {
	var updatedTargets []*model.Timestamp
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
	ts *model.Timestamp,
	targets []*model.Timestamp,
	values []interface{},
) ([]jsonType, errors.OrtooError) {
	errs := &errors.MultipleOrtooErrors{}
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
			_ = errs.Append(errors.DatatypeNoTarget.New(its.GetLogger(), t.ToString()))
		}
	}
	return delTypes, errs.Return()
}

func (its *jsonArray) getValue() types.JSONValue {
	return its.GetAsJSONCompatible()
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
		if ov1.getOrderTime().Compare(model.OldestTimestamp()) == 0 &&
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

func (its *jsonArray) GetBase() iface.BaseDatatype {
	return its.getCommon().base
}

func (its *jsonArray) SetBase(base iface.BaseDatatype) {
	its.getCommon().SetBase(base)
}

func (its *jsonArray) CloneSnapshot() iface.Snapshot {
	panic("Implement me")
}

// GetAsJSONCompatible returns an interface type that contains all live objects.
func (its *jsonArray) GetAsJSONCompatible() interface{} {
	var list = make([]interface{}, 0)
	n := its.listSnapshot.head.getNextLive()
	for n != nil {
		if !n.isTomb() {
			switch cast := n.getTimedType().(type) {
			case *jsonObject:
				list = append(list, cast.GetAsJSONCompatible())
			case *jsonElement:
				list = append(list, cast.getValue())
			case *jsonArray:
				list = append(list, cast.GetAsJSONCompatible())
			}
		}
		n = n.getNextLive()
	}
	return list
}
