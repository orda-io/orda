package orda

import (
	"github.com/orda-io/orda/client/pkg/model"
)

type orderedType interface {
	timedType
	getOrderTime() *model.Timestamp
	setOrderTime(ts *model.Timestamp)
	getPrev() orderedType
	setPrev(n orderedType)
	getNext() orderedType
	setNext(n orderedType)
	insertNext(n orderedType)
	getNextLive() orderedType
	getTimedType() timedType
	setTimedType(tt timedType)
	hash() string
	marshal() *marshaledNode
}

type orderedNode struct {
	timedType
	O    *model.Timestamp
	prev orderedType
	next orderedType
}

func newHead() *orderedNode {
	return &orderedNode{
		timedType: newTimedNode(nil, nil),
		O:         model.OldestTimestamp(),
		prev:      nil,
		next:      nil,
	}
}

func (its *orderedNode) getOrderTime() *model.Timestamp {
	return its.O
}

func (its *orderedNode) setOrderTime(ts *model.Timestamp) {
	its.O = ts
}

func (its *orderedNode) hash() string {
	return its.O.Hash()
}

func (its *orderedNode) getPrev() orderedType {
	return its.prev
}

func (its *orderedNode) setPrev(n orderedType) {
	its.prev = n
}

func (its *orderedNode) getNext() orderedType {
	return its.next
}

func (its *orderedNode) setNext(n orderedType) {
	its.next = n
}

func (its *orderedNode) insertNext(n orderedType) {
	oldNext := its.next
	its.next = n
	n.setPrev(its)
	n.setNext(oldNext)
	if oldNext != nil {
		oldNext.setPrev(n)
	}
}

func (its *orderedNode) getNextLive() orderedType {
	ret := its.next
	for ret != nil {
		if !ret.isTomb() {
			return ret
		}
		ret = ret.getNext()
	}
	return nil
}

func (its *orderedNode) getTimedType() timedType {
	return its.timedType
}

func (its *orderedNode) setTimedType(tt timedType) {
	its.timedType = tt
}
