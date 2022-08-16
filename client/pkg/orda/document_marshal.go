package orda

import (
	"github.com/orda-io/orda/client/pkg/model"
)

type marshaledDocument struct {
	NodeMap []*marshaledJSONType `json:"nm"`
}

type unmarshalAssistant struct {
	tsMap  map[string]*model.Timestamp
	common *jsonCommon
}

type marshalKeyJSONType string

const (
	marshalKeyJSONElement marshalKeyJSONType = "E"
	marshalKeyJSONObject  marshalKeyJSONType = "O"
	marshalKeyJSONArray   marshalKeyJSONType = "A"
)

type marshaledJSONType struct {
	C *model.Timestamp     `json:"c,omitempty"` // jsonPrimitive.C
	T marshalKeyJSONType   `json:"t,omitempty"` // type; "E": jsonElement, "O": jsonObject, "A": jsonArray
	P *model.Timestamp     `json:"p,omitempty"` // jsonPrimitive.parent's C
	D *model.Timestamp     `json:"d,omitempty"` // jsonPrimitive.D
	E interface{}          `json:"e,omitempty"` // for jsonElement
	A *marshaledJSONArray  `json:"a,omitempty"` // for jsonArray
	O *marshaledJSONObject `json:"o,omitempty"` // for jsonObject
}

type marshaledOrderedType [2]*model.Timestamp

type marshaledJSONObject struct {
	M map[string]*model.Timestamp `json:"m"` // hashmapSnapshot.Map
	S int                         `json:"s"` // hashmapSnapshot.Size
}

type marshaledJSONArray struct {
	N []marshaledOrderedType `json:"n"` // an array of node
	S int                    `json:"s"` // size of a list
}

// unifyTimestamp is used to unify timestamps. This must be called when timestamps of any jsonType are set.
func (its *unmarshalAssistant) unifyTimestamp(ts *model.Timestamp) *model.Timestamp {
	if ts == nil {
		return nil
	}
	if existing, ok := its.tsMap[ts.Hash()]; ok {
		return existing
	}
	its.tsMap[ts.Hash()] = ts
	return ts
}

func newMarshaledDocument() *marshaledDocument {
	return &marshaledDocument{
		NodeMap: nil,
	}
}

func (its *marshaledJSONType) unmarshalAsJSONType(assistant *unmarshalAssistant) jsonType {
	jsonType := &jsonPrimitive{
		common: assistant.common,
		C:      assistant.unifyTimestamp(its.C),
		D:      assistant.unifyTimestamp(its.D),
	}
	switch its.T {
	case marshalKeyJSONElement:
		return &jsonElement{
			jsonType: jsonType,
		}
	case marshalKeyJSONObject:
		return &jsonObject{
			jsonType: jsonType,
		}
	case marshalKeyJSONArray:
		return &jsonArray{
			jsonType: jsonType,
		}
	}
	return nil
}
