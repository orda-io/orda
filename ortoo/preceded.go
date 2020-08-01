package ortoo

import (
	"fmt"
	"github.com/knowhunger/ortoo/ortoo/model"
	"github.com/knowhunger/ortoo/ortoo/types"
	"strings"
)

type precededType interface {
	timedType
	getPrecedence() *model.Timestamp
	setPrecedence(ts *model.Timestamp)
}

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

func (its *precededNode) getPrecedence() *model.Timestamp {
	return its.P
}

func (its *precededNode) setPrecedence(ts *model.Timestamp) {
	its.P = ts
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
