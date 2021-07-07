package orda

import (
	"fmt"
	"github.com/orda-io/orda/pkg/model"
	"github.com/orda-io/orda/pkg/types"
)

// timedType is the most primitive type which allows to store a timestamp.
type timedType interface {
	getValue() types.JSONValue
	setValue(v types.JSONValue)
	// getTime and setTime are used when the timestamp is used to resolve conflict.
	// It can be overridden.
	getTime() *model.Timestamp
	setTime(ts *model.Timestamp)
	makeTomb(ts *model.Timestamp)
	isTomb() bool
	String() string
}

type timedNode struct {
	V types.JSONValue  `json:"v"`
	T *model.Timestamp `json:"t"`
}

func newTimedNode(v types.JSONValue, t *model.Timestamp) timedType {
	return &timedNode{
		V: v,
		T: t,
	}
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

// this is for ordaMap and list
func (its *timedNode) makeTomb(ts *model.Timestamp) {
	its.T = ts
	its.V = nil
}

func (its *timedNode) isTomb() bool {
	return its.V == nil
}

func (its *timedNode) String() string {
	if its.V == nil {
		return fmt.Sprintf("Φ|%s", its.T.ToString())
	}
	return fmt.Sprintf("TV[%v|C%s]", its.V, its.T.ToString())
}
