package commons

import "fmt"

type IntCounter struct {
	BaseDataType
	value int32
}

func NewIntCounter() *IntCounter {
	return &IntCounter{
		BaseDataType: BaseDataType{
			id:     newDatatypeID(),
			opID:   newOperationID(),
			typeOf: TypeIntCounter,
			state:  StateLocallyExisted,
		},
		value: 0,
	}
}

func (c *IntCounter) Increase() {
	op := newIncreaseOperation()
	c.execute(op)
}

type increaseOperation struct {
	delta int32
	*operation
}

func newIncreaseOperation() *increaseOperation {
	return &increaseOperation{
		delta: 1,
	}
}

func (c *increaseOperation) executeLocal() {
	fmt.Println("increaseOperation executeLocal")
}
