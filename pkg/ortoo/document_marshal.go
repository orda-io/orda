package ortoo

import (
	"encoding/json"
	"github.com/TylerBrock/colorjson"
	"github.com/knowhunger/ortoo/pkg/log"
	"github.com/knowhunger/ortoo/pkg/model"
)

type marshaledDocument struct {
	NodeMap []*marshaledJSONType `json:"nm"`
	// Cemetery []*model.Timestamp   `json:"ce"` // store the hash of all deleted jsonType
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
	P *model.Timestamp     `json:"p,omitempty"` // jsonPrimitive.parent's C
	K *model.Timestamp     `json:"k,omitempty"` // jsonPrimitive.K
	C *model.Timestamp     `json:"c,omitempty"` // jsonPrimitive.C
	D *model.Timestamp     `json:"d,omitempty"` // jsonPrimitive.D
	T marshalKeyJSONType   `json:"t,omitempty"` // type; "E": jsonElement, "O": jsonObject, "A": jsonArray
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
	} else {
		its.tsMap[ts.Hash()] = ts
		return ts
	}
}

func newMarshaledDocument() *marshaledDocument {
	return &marshaledDocument{
		NodeMap: nil,
		// Cemetery: nil,
	}
}

// //////////////////////////////
// Marshal functions
// //////////////////////////////

// MarshalJSON returns marshaledDocument.
func (its *jsonObject) MarshalJSON() ([]byte, error) {
	marshalDoc := newMarshaledDocument()
	for _, v := range its.getRoot().nodeMap {
		marshaled := v.marshal()
		marshalDoc.NodeMap = append(marshalDoc.NodeMap, marshaled)
	}
	// for _, v := range its.getRoot().cemetery {
	// 	marshalDoc.Cemetery = append(marshalDoc.Cemetery, v.getCreateTime())
	// }
	// printMarshalDoc(marshalDoc)
	return json.Marshal(marshalDoc)
}

func (its *jsonPrimitive) marshal() *marshaledJSONType {
	var p *model.Timestamp = nil
	if its.parent != nil {
		p = its.parent.getCreateTime()
	}
	return &marshaledJSONType{
		P: p,
		C: its.C,
		D: its.D,
	}
}

func (its *jsonElement) marshal() *marshaledJSONType {
	forMarshal := its.jsonType.marshal()
	forMarshal.T = marshalKeyJSONElement
	forMarshal.E = its.getValue()
	return forMarshal
}

func (its *jsonObject) marshal() *marshaledJSONType {
	marshal := its.jsonType.marshal()
	marshal.T = marshalKeyJSONObject
	value := &marshaledJSONObject{
		S: its.Size,
		M: make(map[string]*model.Timestamp),
	}
	for k, v := range its.hashMapSnapshot.Map {
		jt := v.(jsonType)
		value.M[k] = jt.getCreateTime() // store only create timestamp
	}
	marshal.O = value
	return marshal
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
			mot = [2]*model.Timestamp{n.getOrderTime()}
		} else {
			mot = [2]*model.Timestamp{n.getOrderTime(), jt.getCreateTime()}
		}

		marshaledJA.N = append(marshaledJA.N, mot)
		n = n.getNext()
	}
	marshal.A = marshaledJA
	return marshal
}

func printMarshalDoc(doc *marshaledDocument) {
	f := colorjson.NewFormatter()
	f.Indent = 2
	f.DisabledColor = true
	m, _ := json.Marshal(doc)
	var obj map[string]interface{}
	_ = json.Unmarshal(m, &obj)
	s, _ := f.Marshal(obj)
	log.Logger.Infof("%v", string(s))
}

// //////////////////////////////
// Unmarshal functions
// //////////////////////////////

func (its *jsonObject) UnmarshalJSON(bytes []byte) error {
	var forUnmarshal marshaledDocument

	if err := json.Unmarshal(bytes, &forUnmarshal); err != nil {
		return log.OrtooError(err)
	}

	// printMarshalDoc(&forUnmarshal)

	assistant := &unmarshalAssistant{
		tsMap: make(map[string]*model.Timestamp),
		common: &jsonCommon{
			root:     its,
			base:     its.base,
			nodeMap:  make(map[string]jsonType),
			cemetery: make(map[string]jsonType),
		},
	}
	// make all skeleton jsonTypes "in advance"
	for _, v := range forUnmarshal.NodeMap {
		jt := v.unmarshalAsJSONType(&forUnmarshal, assistant)
		if jt.getCreateTime().Compare(model.OldestTimestamp()) == 0 {
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

func (its *jsonPrimitive) unmarshal(marshaled *marshaledJSONType, assistant *unmarshalAssistant) {
	// do nothing
}

func (its *jsonElement) unmarshal(marshaled *marshaledJSONType, assistant *unmarshalAssistant) {
	its.V = marshaled.E
}

func (its *jsonObject) unmarshal(marshaled *marshaledJSONType, assistant *unmarshalAssistant) {
	marshaledJO := marshaled.O
	its.hashMapSnapshot = &hashMapSnapshot{
		base: its.getBase(),
		Map:  make(map[string]timedType),
		Size: marshaledJO.S,
	}
	for k, ts := range marshaledJO.M {
		realInstance, _ := its.findJSONType(ts)
		its.hashMapSnapshot.Map[k] = realInstance
	}
}

func (its *jsonArray) unmarshal(marshaled *marshaledJSONType, assistant *unmarshalAssistant) {
	marshaledJA := marshaled.A
	its.listSnapshot = newListSnapshot(its.getBase())
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

func (its *marshaledJSONType) unmarshalAsJSONPrimitive(doc *marshaledDocument, assistant *unmarshalAssistant) *jsonPrimitive {
	return &jsonPrimitive{
		common: assistant.common,
		C:      assistant.unifyTimestamp(its.C),
		D:      assistant.unifyTimestamp(its.D),
	}
}

func (its *marshaledJSONType) unmarshalAsJSONType(doc *marshaledDocument, assistant *unmarshalAssistant) jsonType {
	switch its.T {
	case marshalKeyJSONElement:
		return &jsonElement{
			jsonType: its.unmarshalAsJSONPrimitive(doc, assistant),
		}
	case marshalKeyJSONObject:
		return &jsonObject{
			jsonType: its.unmarshalAsJSONPrimitive(doc, assistant),
		}
	case marshalKeyJSONArray:
		return &jsonArray{
			jsonType: its.unmarshalAsJSONPrimitive(doc, assistant),
		}
	}
	return nil
}
