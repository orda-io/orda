package commons

import (
	"github.com/golang/protobuf/proto"
	"github.com/knowhunger/ortoo/client"
	"github.com/knowhunger/ortoo/commons/internal/datatypes"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
)

type IntCounter interface {
	datatypes.PublicWiredDatatypeInterface
	IntCounterTransaction
	DoTransaction(tag string, transFunc func(intCounter IntCounterTransaction) error) error
}

type IntCounterTransaction interface {
	Get() int32
	Increase() (int32, error)
	IncreaseBy(delta int32) (int32, error)
}

type intCounter struct {
	*datatypes.WiredDatatypeImpl
	ctx         *intCounterContext
	trnxContext *datatypes.TransactionContext
	trnxManager *datatypes.TransactionManager
}

type intCounterContext struct {
	value int32
}

func NewIntCounter(c client.Client, w datatypes.Wire) (IntCounter, error) {
	wiredDatatype, err := datatypes.NewWiredDataType(model.TypeDatatype_INT_COUNTER, w)
	if err != nil {
		return nil, log.OrtooError(err, "fail to create int counter due to wiredDatatype")
	}
	ctx := &intCounterContext{
		value: 0,
	}
	trnxMgr := datatypes.NewTransactionManager()
	intCounter := &intCounter{
		WiredDatatypeImpl: wiredDatatype,
		ctx:               ctx,
		trnxManager:       trnxMgr,
		trnxContext:       nil,
	}
	intCounter.SetOperationExecuter(intCounter)
	return intCounter, nil
}

func (c *intCounter) Get() int32 {
	return c.ctx.value
}

func (c *intCounter) Increase() (int32, error) {
	return c.IncreaseBy(1)
}

func (c *intCounter) IncreaseBy(delta int32) (int32, error) {
	op := model.NewIncreaseOperation(delta)
	trnxCtx := c.trnxManager.BeginTransaction("", c.trnxContext)
	defer c.trnxManager.EndTransaction(trnxCtx)
	ret, err := c.ExecuteWired(op)
	if err != nil {
		return 0, log.OrtooError(err, "")
	}
	return ret.(int32), nil
}

func (c *intCounterContext) increaseCommon(delta int32) int32 {
	log.Logger.Info("increaseCommon")
	c.value = c.value + delta
	return c.value
}

//ExecuteLocal is the
func (c *intCounter) ExecuteLocal(op interface{}) (interface{}, error) {
	iop := op.(*model.IncreaseOperation)
	log.Logger.Info("delta:", proto.MarshalTextString(iop))
	return c.ctx.increaseCommon(iop.Delta), nil
	//return nil, nil
}

func (c *intCounter) ExecuteRemote(op interface{}) (interface{}, error) {
	iop := op.(*model.IncreaseOperation)
	return c.ctx.increaseCommon(iop.Delta), nil
}

func (c *intCounter) GetWired() datatypes.WiredDatatype {
	return c.WiredDatatypeImpl
}

func (c *intCounter) DoTransaction(tag string, transFunc func(intCounter IntCounterTransaction) error) error {
	ctx := c.trnxManager.BeginTransaction(tag, c.trnxContext)
	defer c.trnxManager.EndTransaction(ctx)
	cc := &intCounter{
		WiredDatatypeImpl: c.WiredDatatypeImpl,
		ctx:               c.ctx,
		trnxManager:       c.trnxManager,
		trnxContext:       ctx,
	}
	err := transFunc(cc)
	if err != nil {

		return log.OrtooError(err, "fail to do the transaction")
	}
	return nil
}
