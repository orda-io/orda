package commons

import (
	"github.com/knowhunger/ortoo/commons/internal/datatypes"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
)

type IntCounter interface {
	Get() int32
	Increase() (int32, error)
	IncreaseBy(delta int32) (int32, error)
}

type intCounterImpl struct {
	*datatypes.WiredDatatypeImpl
	value int32
}

func NewIntCounter(w datatypes.Wire) (IntCounter, error) {
	wiredDatatype, err := datatypes.NewWiredDataType(model.TypeDatatype_INT_COUNTER, w)
	if err != nil {
		return nil, log.OrtooError(err, "fail to create int counter due to wiredDatatype")
	}
	d := &intCounterImpl{
		WiredDatatypeImpl: wiredDatatype,
		value:             0,
	}
	d.SetOperationExecuter(d)
	return d, nil
}

func (c *intCounterImpl) Get() int32 {
	return c.value
}

func (c *intCounterImpl) Increase() (int32, error) {
	return c.IncreaseBy(1)
}

func (c *intCounterImpl) IncreaseBy(delta int32) (int32, error) {
	op := model.NewIncreaseOperation(delta)
	_, err := c.ExecuteWired(c, op)
	return c.value, err
}

func (c *intCounterImpl) increaseCommon(delta int32) int32 {
	c.Info("increaseCommon")
	c.value = c.value + delta
	return c.value
}

//ExecuteLocal is the
func (c *intCounterImpl) ExecuteLocal(op interface{}) (interface{}, error) {
	iop := op.(*model.IncreaseOperation)
	log.Logger.Info("delta:", iop)
	return c.increaseCommon(iop.Delta), nil
	//return nil, nil
}

func (c *intCounterImpl) ExecuteRemote(op interface{}) (interface{}, error) {
	iop := op.(*model.IncreaseOperation)
	return c.increaseCommon(iop.Delta), nil
}
