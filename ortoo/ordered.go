package ortoo

type orderedType interface {
	precededType
	getPrev() orderedType
	setPrev(n orderedType)
	getNext() orderedType
	setNext(n orderedType)
	getNextLive() orderedType
	hash() string
	toNodeForMarshal() *nodeForMarshal
	getPrecededType() precededType
}

type orderedNode struct {
	precededType
	next orderedType
	prev orderedType
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
