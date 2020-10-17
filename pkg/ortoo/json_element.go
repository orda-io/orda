package ortoo

import (
	"fmt"
	"github.com/knowhunger/ortoo/pkg/iface"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/pkg/types"
)

// ////////////////////////////////////
//  jsonElement
// ////////////////////////////////////

type jsonElement struct {
	jsonType
	V types.JSONValue
}

func newJSONElement(parent jsonType, value interface{}, ts *model.Timestamp) *jsonElement {
	return &jsonElement{
		jsonType: &jsonPrimitive{
			parent: parent,
			common: parent.getCommon(),
			C:      ts,
		},
		V: value,
	}
}

func (its *jsonElement) getValue() types.JSONValue {
	return its.V
}

func (its *jsonElement) getType() TypeOfJSON {
	return TypeJSONElement
}

func (its *jsonElement) setValue(v types.JSONValue) {
	panic("not used yet")
}

func (its *jsonElement) String() string {
	parent := its.getParent()
	parentTS := "nil"
	if parent != nil {
		parentTS = parent.getCreateTime().ToString()
	}
	value := its.V
	if its.isTomb() {
		value = "#!DELETED"
	}
	return fmt.Sprintf("JE(P%v)[C%v|%v]", parentTS, its.getCreateTime().ToString(), value)
}

func (its *jsonElement) equal(o jsonType) bool {
	if its.getType() != o.getType() {
		return false
	}
	je := o.(*jsonElement)
	if !its.jsonType.equal(je.jsonType) {
		return false
	}

	if its.V != je.V {
		return false
	}
	return true
}

// ///////////////////// methods of iface.Snapshot ///////////////////////////////////////

func (its *jsonElement) CloneSnapshot() iface.Snapshot {
	panic("Implement me")
}

func (its *jsonElement) GetAsJSONCompatible() interface{} {
	return its.getValue()
}
