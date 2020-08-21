package ortoo

import (
	"fmt"
	"github.com/knowhunger/ortoo/ortoo/model"
	"github.com/knowhunger/ortoo/ortoo/types"
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
			common: parent.getRoot(),
			K:      ts,
			P:      ts,
		},
		V: value,
	}
}

func (its *jsonElement) makeTomb(ts *model.Timestamp) bool {
	if its.jsonType.makeTomb(ts) {
		its.addToCemetery(its)
		return true
	}
	return false
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
		parentTS = parent.getKey().ToString()
	}
	value := its.V
	if its.isTomb() {
		value = "#!DELETED"
	}
	return fmt.Sprintf("JE(P%v)[T%v|%v]", parentTS, its.getKey().ToString(), value)
}

func (its *jsonElement) makeTombAsChild(ts *model.Timestamp) bool {
	if its.jsonType.makeTombAsChild(ts) {
		its.addToCemetery(its)
		return true
	}
	return false
}
