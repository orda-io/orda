package commons

import (
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
)

type IntCounter interface {
	Get() int32
	Increase() (int32, error)
	IncreaseBy(delta int32) (int32, error)
}

type IntCounterImpl struct {
	*WiredDatatypeT
	value int32
}

func NewIntCounter(w wire) (*IntCounterImpl, error) {
	wiredDatatype, err := newWiredDataType(TypeIntCounter, w)
	if err != nil {
		return nil, log.OrtooError(err, "fail to create int counter due to wiredDatatype")
	}
	d := &IntCounterImpl{
		WiredDatatypeT: wiredDatatype,
		value:          0,
	}
	d.super = d
	return d, nil
}

func (c *IntCounterImpl) Get() int32 {
	return c.value
}

func (c *IntCounterImpl) Increase() (int32, error) {
	return c.IncreaseBy(1)
}

func (c *IntCounterImpl) IncreaseBy(delta int32) (int32, error) {
	op := model.NewIncreaseOperation(delta)
	_, err := c.executeWired(c, op)
	return c.value, err
}

func (c *IntCounterImpl) increaseCommon(delta int32) int32 {
	c.Info("increaseCommon")
	c.value = c.value + delta
	return c.value
}

func (c *IntCounterImpl) ExecuteLocal(op interface{}) (interface{}, error) {
	iop := op.(*model.IncreaseOperation)
	log.Logger.Info("delta:", iop)
	return c.increaseCommon(iop.Delta), nil
	//return nil, nil
}

func (c *IntCounterImpl) ExecuteRemote(op interface{}) (interface{}, error) {
	iop := op.(*model.IncreaseOperation)
	return c.increaseCommon(iop.Delta), nil
}
