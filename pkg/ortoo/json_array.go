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
			K:      ts,
			P:      ts,
		},
		listSnapshot: newListSnapshot(base),
	}
}

func (its *jsonArray) makeTomb(ts *model.Timestamp) bool {
	if its.jsonType.makeTomb(ts) {
		its.addToCemetery(its)
		// n := its.head.getNext()
		// for n != nil {
		// 	cast := n.getPrecededType().(jsonType)
		// 	cast.makeTombAsChild(ts)
		// 	n = n.getNextLive()
		// }
		return true
	}
	return false
}

func (its *jsonArray) arrayInsertCommon(
	pos int, // in the case of the local insert
	target *model.Timestamp, // in the case of the remote insert
	ts *model.Timestamp,
	values ...interface{},
) (*model.Timestamp, []interface{}, errors.OrtooError) {
	var inserted []precededType
	for _, v := range values {
		element := its.createJSONType(its, v, ts)
		inserted = append(inserted, element)
		its.addToNodeMap(element)
	}
	if target == nil { // InsertLocal
		return its.listSnapshot.insertLocalWithPrecededTypes(pos, inserted...)
	} // InsertRemote
	its.listSnapshot.insertRemoteWithPrecededTypes(target, ts, inserted...)
	return nil, nil, nil
}

func (its *jsonArray) arrayDeleteLocal(pos, numOfNodes int, ts *model.Timestamp) ([]*model.Timestamp, []interface{}, errors.OrtooError) {
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

func (its *jsonArray) arrayDeleteRemote(targets []*model.Timestamp, ts *model.Timestamp) (errs []errors.OrtooError) {

	for _, t := range targets {
		if j, ok := its.findJSONPrimitive(t); ok {
			if !j.isTomb() {
				j.makeTomb(ts)
				its.size--
			} else { // concurrent deletes
				if j.getPrecedence().Compare(ts) < 0 {
					j.setPrecedence(ts)
				}
			}
		} else {
			errs = append(errs, errors.ErrDatatypeNoTarget.New(its.getLogger(), t.ToString()))
		}
	}
	return errs
}

func (its *jsonArray) arrayUpdateLocal(pos int, ts *model.Timestamp, values ...interface{}) errors.OrtooError {
	// if err := its.validateRange(pos, len(values)); err != nil {
	// 	return err
	// }
	// var updated []*model.Timestamp
	// n := its.findNthTarget(pos+1)
	// for _, v := range values {
	// 	old := n.getPrecededType()
	// 	new := its.createJSONType(its, v, ts)
	// 	new.setKey(old.getKey())
	// 	switch cast := old.(type) {
	// 	case *jsonElement:
	// 		its.removeFromNodeMap(cast)
	// 	case *jsonObject:
	// 		its.addToCemetery(cast)
	// 	case *jsonArray:
	// 		its.addToCemetery(cast)
	// 	}
	// 	n.setPrecededType(new)
	// 	its.addToNodeMap(new)
	// 	n = n.getNextLive()
	// }
	// for n != nil {
	// 	if !n.isTomb() {
	// 		switch cast := n.getPrecededType().(type) {
	// 		case *jsonObject:
	// 			// list = append(list, cast.GetAsJSONCompatible())
	// 		case *jsonElement:
	// 			// list = append(list, cast.getValue())
	// 		case *jsonArray:
	// 			// list = append(list, cast.GetAsJSONCompatible())
	// 		}
	// 	}
	// 	n = n.getNextLive()
	// }
	// orderedType := its.findNthTarget(op.Pos + 1)
	// for i := 0; i < len(op.C.V); i++ {
	//
	// }
	return nil
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
		parentTS = parent.getKey().ToString()
	}
	return fmt.Sprintf("JA(%v)[T%v|V%v", parentTS, its.getKey().ToString(), its.listSnapshot.String())
}

// GetAsJSONCompatible returns an interface type that contains all live objects.
func (its *jsonArray) GetAsJSONCompatible() interface{} {
	var list []interface{}
	n := its.listSnapshot.head.getNextLive()
	for n != nil {
		if !n.isTomb() {
			switch cast := n.getPrecededType().(type) {
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
