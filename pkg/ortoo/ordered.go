package ortoo

import "github.com/knowhunger/ortoo/pkg/model"

type orderedType interface {
	precededType
	getPrev() orderedType
	setPrev(n orderedType)
	getNext() orderedType
	setNext(n orderedType)
	getNextLive() orderedType
	getPrecededType() precededType
	hash() string
	marshal() *marshaledNode
}

type orderedNode struct {
	precededType
	prev orderedType
	next orderedType
}

func newHead() *orderedNode {
	return &orderedNode{
		precededType: &precededNode{
			timedType: &timedNode{
				V: nil,
				T: model.OldestTimestamp,
			},
			P: nil,
		},
		prev: nil,
		next: nil,
	}
}

func (its *orderedNode) hash() string {
	return its.getTime().Hash()
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

func (its *orderedNode) getNextLive() orderedType {
	ret := its.next
	for ret != nil {
		if ret.getValue() != nil {
			return ret
		}
		ret = ret.getNext()
	}
	return nil
}

func (its *orderedNode) getPrecededType() precededType {
	return its.precededType
}
