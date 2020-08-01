package ortoo

import (
	"encoding/json"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/ortoo/model"
	"strings"
)

type documentForMarshal struct {
	NodeMap      []string
	Cemetery     []string
	TimestampMap map[string]*model.Timestamp
	MarshalMap   map[string]*jsonPrimitiveForMarshal
}

// // MarshalJSON returns
// func (its *jsonObject) MarshalJSON() ([]byte, error) {
// 	forMarshal := newDocumentForMarshal()
// 	for k, _ := range its.getRoot().nodeMap {
// 		forMarshal.NodeMap = append(forMarshal.NodeMap, k)
// 	}
// 	for k, _ := range its.getRoot().cemetery {
// 		forMarshal.Cemetery = append(forMarshal.Cemetery, k)
// 	}
//
// 	its.marshalDocument(forMarshal)
// 	return nil, nil
// }

func unmarshalFromMap(parent jsonType, m map[string]interface{}) jsonType {
	switch m["T"] {
	case "E":
		return &jsonElement{
			jsonType: &jsonPrimitive{
				parent:  parent,
				root:    nil,
				K:       m["K"].(*model.Timestamp),
				P:       m["P"].(*model.Timestamp),
				deleted: m["D"].(bool),
			},
			V: m["V"],
		}
	case "O":
	case "A":

	}
	return nil
}

type jsonObjectForMarshal struct {
	NodeMap  []string
	Cemetery []string
	Map      map[string]interface{}
}

func newJSONObjectForMarshal() *jsonObjectForMarshal {
	return &jsonObjectForMarshal{
		NodeMap:  nil,
		Cemetery: nil,
		Map:      make(map[string]interface{}),
	}
}

func (its *jsonObject) UnmarshalJSON(bytes []byte) error {
	var forUnmarshal jsonObjectForMarshal
	if err := json.Unmarshal(bytes, &forUnmarshal); err != nil {
		return log.OrtooError(err)
	}
	for k, v := range forUnmarshal.Map {
		log.Logger.Infof("%s: %v", k, v)
		unmarshalFromMap(its, v.(map[string]interface{}))
		// its.Map[k] =v
	}
	return nil
}

type jsonPrimitiveForMarshal struct {
	T string           // type; "E": jsonElement, "O": jsonObject, "A": jsonArray
	K *model.Timestamp // jsonPrimitive.K
	P *model.Timestamp // jsonPrimitive.P
	D bool             // jsonPrimitive.deleted
	V interface{}      // varying depending on jsonElement, jsonObject, jsonArray
}

func (its *jsonPrimitive) toJSONPrimitiveForMarshal() *jsonPrimitiveForMarshal {
	return &jsonPrimitiveForMarshal{
		K: its.K,
		P: its.P,
		D: its.deleted,
	}
}

func (its *jsonElement) toJSONPrimitiveForMarshal() *jsonPrimitiveForMarshal {
	forMarshal := its.jsonType.toJSONPrimitiveForMarshal()
	forMarshal.T = "E"
	forMarshal.V = its.getValue()
	return forMarshal
}

func (its *jsonObject) toJSONPrimitiveForMarshal() *jsonPrimitiveForMarshal {
	forMarshal := its.jsonType.toJSONPrimitiveForMarshal()
	forMarshal.T = "O"
	value := &struct {
		Map  map[string]*model.Timestamp
		Size int
	}{
		Size: its.Size,
		Map:  make(map[string]*model.Timestamp),
	}
	for k, v := range its.hashMapSnapshot.Map {
		jsonP := v.(jsonType)
		value.Map[k] = jsonP.getTime()
	}
	forMarshal.V = value
	return forMarshal
}

func (its *jsonArray) toJSONPrimitiveForMarshal() *jsonPrimitiveForMarshal {
	forMarshal := its.jsonType.toJSONPrimitiveForMarshal()
	forMarshal.T = "A"
	value := &struct {
		Size int
	}{
		Size: its.size,
	}
	forMarshal.V = value

	return forMarshal
}

func (its *jsonPrimitiveForMarshal) fromJSONElementForMarshal() *jsonElement {
	return &jsonElement{
		jsonType: &jsonPrimitive{
			parent: nil,
			root:   nil,
			// K:       its.K,
			// P:       its.P,
			deleted: its.D,
		},
	}
}

func (its *jsonElement) MarshalJSON() ([]byte, error) {
	j := its.toJSONPrimitiveForMarshal()
	return json.Marshal(j)
}

func (its *jsonElement) UnmarshalJSON(bytes []byte) error {
	// forMarshal := jsonPrimitiveForMarshal{}
	// if err := json.Unmarshal(bytes, &forMarshal); err != nil {
	// 	return log.OrtooError(err)
	// }
	//
	// its.V = forMarshal.V
	// its.jsonType = &jsonPrimitive{
	// 	parent:  nil,
	// 	root:    nil,
	// 	K:       forMarshal.K,
	// 	P:       forMarshal.P,
	// 	deleted: false,
	// }
	return nil
}

func newDocumentForMarshal() *documentForMarshal {
	return &documentForMarshal{
		NodeMap:      nil,
		Cemetery:     nil,
		TimestampMap: make(map[string]*model.Timestamp),
		MarshalMap:   make(map[string]*jsonPrimitiveForMarshal),
	}
}

func (its *jsonElement) marshalDocument(forMarshal *documentForMarshal) {
	k := its.jsonType.getTime()
	if k != nil {
		forMarshal.TimestampMap[k.Hash()] = k
	}
	p := its.jsonType.getPrecedence()
	if p != nil {
		forMarshal.TimestampMap[p.Hash()] = p
	}
	m := its.toJSONPrimitiveForMarshal()
	forMarshal.MarshalMap[m.K.Hash()] = m
}

func (its *jsonArray) marshalDocument(forMarshal *documentForMarshal) {
	var sb strings.Builder
	n := its.head
	for n != nil {
		if n.getPrev() != nil {
			sb.WriteString(n.getPrev().hash())
		}
		if n.getNext() != nil {
			sb.WriteString(n.getNext().hash())
		}
		// switch cast:= n.(type) {
		// case *jsonObject:
		// case *jsonElement:
		// case *jsonArray:
		// }
		n = n.getNext()
	}
}

func (its *jsonObject) marshalDocument(forMarshal *documentForMarshal) {
	var sb strings.Builder

	for k, v := range its.Map {
		if v != nil {
			switch cast := v.(type) {
			case *jsonObject:
				cast.marshalDocument(forMarshal)
			case *jsonElement:
				cast.marshalDocument(forMarshal)
				// kts := cast.jsonType.getKey()
				// pts := cast.jsonType.getTime()
				//
				sb.WriteString(k)
				sb.WriteString(" ")
				j, err := json.Marshal(cast)
				if err != nil {
					return
				}
				sb.WriteString(string(j))
			case *jsonArray:
				cast.marshalDocument(forMarshal)

			}
		}
		sb.WriteString("\n")
	}
	log.Logger.Infof("%v", sb.String())
	// b, err := json.Marshal(forMarshal)
	// if err != nil {
	// 	return
	// }
	return
}
