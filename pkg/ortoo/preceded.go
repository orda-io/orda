package ortoo

import (
	"fmt"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/pkg/types"
	"strings"
)

type precededType interface {
	timedType
	// getKey and setKey are used when the timestamp is contextually used as an immutable key.
	getKey() *model.Timestamp
	setKey(ts *model.Timestamp)

	getPrecedence() *model.Timestamp
	setPrecedence(ts *model.Timestamp)
}

// precededNode implements timedType and precededType, which is used in list
type precededNode struct {
	timedType
	P *model.Timestamp
}

func newPrecededNode(v types.JSONValue, t *model.Timestamp, p *model.Timestamp) *precededNode {
	return &precededNode{
		timedType: &timedNode{
			V: v,
			T: t,
		},
		P: p,
	}
}

// ///////////////////// methods of precededType ///////////////////////////////////

func (its *precededNode) getKey() *model.Timestamp {
	return its.timedType.getTime()
}

func (its *precededNode) setKey(ts *model.Timestamp) {
	its.timedType.setTime(ts)
}

func (its *precededNode) getPrecedence() *model.Timestamp {
	return its.P
}

func (its *precededNode) setPrecedence(ts *model.Timestamp) {
	its.P = ts
}

// ///////////////////// methods of timedType ///////////////////////////////////

func (its *precededNode) getTime() *model.Timestamp {
	if its.P != nil {
		return its.P
	}
	return its.timedType.getTime()
}

// override makeTomb() for list
func (its *precededNode) makeTomb(ts *model.Timestamp) bool {
	its.setValue(nil)
	its.P = ts
	return true
}

func (its *precededNode) isTomb() bool {
	return its.getValue() == nil && its.P != nil
}

// ///////////////////// other methods ///////////////////////////////////

func (its *precededNode) String() string {
	var sb strings.Builder
	sb.WriteString(its.getTime().ToString())
	if its.P != nil {
		sb.WriteString(its.P.ToString())
	}
	if its.getValue() == nil {
		sb.WriteString(":DELETED")
	} else {
		_, _ = fmt.Fprintf(&sb, ":%v", its.getValue())
	}

	return sb.String()
}
