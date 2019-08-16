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
	IntCounterInTransaction
	DoTransaction(tag string, transFunc func(intCounter IntCounterInTransaction) error) error
}

type IntCounterInTransaction interface {
	Get() int32
	Increase() (int32, error)
	IncreaseBy(delta int32) (int32, error)
}

type intCounter struct {
	*datatypes.TransactionManager
	ctx            *intCounterContext
	transactionCtx *datatypes.TransactionContext
}

func NewIntCounter(c client.Client, w datatypes.Wire) (IntCounter, error) {
	ctx := &intCounterContext{
		value: 0,
	}
	transactionMgr, err := datatypes.NewTransactionManager(model.TypeDatatype_INT_COUNTER, w)
	if err != nil {
		return nil, log.OrtooError(err, "fail to create transaction manager")
	}
	intCounter := &intCounter{
		TransactionManager: transactionMgr,
		ctx:                ctx,
		transactionCtx:     nil,
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
	ret, err := c.Execute(c.transactionCtx, op)
	if err != nil {
		return 0, log.OrtooError(err, "fail to execute operation")
	}
	return ret.(int32), nil
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

func (c *intCounter) DoTransaction(tag string, transFunc func(intCounter IntCounterInTransaction) error) error {
	log.Logger.Infof("Before BeginTransaction:%s", tag)
	transactionCtx, err := c.BeginTransaction(tag, c.transactionCtx, true)
	log.Logger.Infof("End BeginTransaction:%s", tag)
	defer c.EndTransaction(transactionCtx, true)
	clone := &intCounter{
		TransactionManager: c.TransactionManager,
		ctx:                c.ctx,
		transactionCtx:     transactionCtx,
	}
	err = transFunc(clone)
	if err != nil {
		c.SetTransactionFail()
		return log.OrtooError(err, "fail to do the transaction")
	}
	return nil
}

type intCounterContext struct {
	value int32
}

func (c *intCounterContext) increaseCommon(delta int32) int32 {
	log.Logger.Info("increaseCommon")
	c.value = c.value + delta
	return c.value
}
