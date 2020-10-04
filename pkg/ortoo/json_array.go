package ortoo

import (
	"fmt"
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/internal/datatypes"
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

func newJSONArray(base *datatypes.BaseDatatype, parent jsonType, ts *model.Timestamp) *jsonArray {
	return &jsonArray{
		jsonType: &jsonPrimitive{
			parent: parent,
			common: parent.getRoot(),
			T:      ts,
		},
		listSnapshot: newListSnapshot(base),
	}
}

func (its *jsonArray) getJSONType(pos int) (jsonType, errors.OrtooError) {
	tt, err := its.findTimedType(pos)
	if err != nil {
		return nil, err
	}
	return tt.(jsonType), nil
}

func (its *jsonArray) insertCommon(
	pos int, // in the case of the local insert
	target *model.Timestamp, // in the case of the remote insert
	ts *model.Timestamp,
	values ...interface{},
) (*model.Timestamp, []interface{}, errors.OrtooError) {
	var inserted []timedType
	for _, v := range values {
		element := its.createJSONType(its, v, ts)
		inserted = append(inserted, element)
		its.addToNodeMap(element)
	}
	if target == nil {
		return its.listSnapshot.insertLocalWithTimedTypes(pos, inserted...)
	}
	return nil, nil, its.listSnapshot.insertRemoteWithTimedTypes(target, ts, inserted...)
}

func (its *jsonArray) deleteLocal(
	pos int,
	numOfNodes int,
	ts *model.Timestamp,
) ([]*model.Timestamp, []interface{}, errors.OrtooError) {
	targets, values, err := its.listSnapshot.deleteLocal(pos, numOfNodes, ts)
	if err != nil {
		return nil, nil, err
	}
	for _, v := range targets {
		if jt, ok := its.findJSONPrimitive(v); ok {
			its.addToCemetery(jt)
		}
	}
	return targets, values, err
}

func (its *jsonArray) deleteRemote(targets []*model.Timestamp, ts *model.Timestamp) (errs []errors.OrtooError) {
	deleted, errs := its.listSnapshot.deleteRemote(targets, ts)
	for _, t := range deleted {
		its.addToCemetery(t.(jsonType))
	}
	return errs
}

func (its *jsonArray) updateLocal(
	pos int,
	ts *model.Timestamp,
	values ...interface{},
) ([]*model.Timestamp, errors.OrtooError) {
	if err := its.validateRange(pos, len(values)); err != nil {
		return nil, err
	}
	var updatedTargets []*model.Timestamp
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
		target = target.getNextLive()
	}
	return updatedTargets, nil
}

func (its *jsonArray) updateRemote(
	ts *model.Timestamp,
	targets []*model.Timestamp,
	values []interface{},
) (errs []errors.OrtooError) {

	for i, t := range targets {
		newOne := its.createJSONType(its, values[i], ts)
		its.addToNodeMap(newOne)
		// thisTS := ts.GetAndNextDelimiter()
		if node, ok := its.Map[t.Hash()]; ok {
			var deleted, updated jsonType
			oldOne := node.getTimedType().(jsonType)
			if !node.isTomb() {
				if node.getTime().Compare(newOne.getKeyTime()) < 0 {
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
			its.funeral(deleted, updated.getKeyTime())
		} else {
			errs = append(errs, errors.ErrDatatypeNoTarget.New(its.base.Logger, t.ToString()))
		}
	}
	return
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
		parentTS = parent.getKeyTime().ToString()
	}
	return fmt.Sprintf("JA(%v)[T%v|V%v", parentTS, its.getKeyTime().ToString(), its.listSnapshot.String())
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
