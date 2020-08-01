package ortoo

import (
	"fmt"
	"github.com/knowhunger/ortoo/ortoo/model"
	"github.com/knowhunger/ortoo/ortoo/types"
)

type timedType interface {
	getValue() types.JSONValue
	setValue(v types.JSONValue)
	getTime() *model.Timestamp
	setTime(ts *model.Timestamp)
	makeTomb(ts *model.Timestamp) bool
	isTomb() bool
	String() string
}

type timedNode struct {
	V types.JSONValue  `json:"v"`
	T *model.Timestamp `json:"t"`
}

func (its *timedNode) getValue() types.JSONValue {
	return its.V
}

func (its *timedNode) setValue(v types.JSONValue) {
	its.V = v
}

func (its *timedNode) getTime() *model.Timestamp {
	return its.T
}

func (its *timedNode) setTime(ts *model.Timestamp) {
	its.T = ts
}

// this is for hash_map
func (its *timedNode) makeTomb(ts *model.Timestamp) bool {
	its.V = nil
	its.T = ts
	return true
}

func (its *timedNode) isTomb() bool {
	if its.V == nil {
		return true
	}
	return false
}

func (its *timedNode) String() string {
	if its.V == nil {
		return fmt.Sprintf("Φ|%s", its.T.ToString())
	}
	return fmt.Sprintf("TV[%v|T%s]", its.V, its.T.ToString())
}
