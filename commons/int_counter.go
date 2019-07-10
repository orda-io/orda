package commons

import (
	. "github.com/knowhunger/ortoo/commons/utils"
)

type IntCounter struct {
	*WiredDatatypeT
	value int32
}

func NewIntCounter(w wire) *IntCounter {
	d := &IntCounter{
		WiredDatatypeT: newWiredDataType(TypeIntCounter, w),
		value:          0,
	}
	d.super = d
	return d
}

func (c *IntCounter) Increase() (int32, error) {
	return c.IncreaseBy(1)
}

func (c *IntCounter) IncreaseBy(delta int32) (int32, error) {
	op := newIncreaseOperation(delta)
	_, err := execute(c, op)
	return c.value, err
}

func (c *IntCounter) Get() int32 {
	return c.value
}

func (c *IntCounter) increaseLocal(delta int32) int32 {
	Log.Info("increaseLocal")
	c.value = c.value + delta
	return c.value
}

type increaseOperation struct {
	delta int32
	*operationT
}

func newIncreaseOperation(delta int32) *increaseOperation {
	return &increaseOperation{
		delta: delta,
		operationT: &operationT{
			typ: OperationTypes.IntCounterIncreaseType,
		},
	}
}

func (i *increaseOperation) executeLocal(datatype interface{}) (interface{}, error) {
	if counter, ok := datatype.(*IntCounter); ok {
		Log.Println("increaseOperation executeLocal")
		return counter.increaseLocal(i.delta), nil
	}
	return nil, OrtooError("operation is called with invalid datatype")
}

func (i *increaseOperation) executeRemote(datatype interface{}) {
	if counter, ok := datatype.(*IntCounter); ok {
		Log.Println("increaseOperation executeRemote")
		counter.increaseLocal(i.delta)
	}
}
