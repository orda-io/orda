package commons

import (
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
	*datatypes.TransactionDatatypeImpl
	snapshot       *intCounterSnapshot
	transactionCtx *datatypes.TransactionContext
}

func NewIntCounter(c client.Client, w datatypes.Wire) (IntCounter, error) {
	snapshot := &intCounterSnapshot{
		value: 0,
	}
	transactionDatatype, err := datatypes.NewTransactionDatatype(model.TypeDatatype_INT_COUNTER, w, snapshot)
	if err != nil {
		return nil, log.OrtooError(err, "fail to create transaction manager")
	}
	intCounter := &intCounter{
		TransactionDatatypeImpl: transactionDatatype,
		snapshot:                snapshot,
		transactionCtx:          nil,
	}
	intCounter.SetOperationExecuter(intCounter)
	return intCounter, nil
}

func (c *intCounter) Get() int32 {
	return c.snapshot.value
}

func (c *intCounter) Increase() (int32, error) {
	return c.IncreaseBy(1)
}

func (c *intCounter) IncreaseBy(delta int32) (int32, error) {
	op := model.NewIncreaseOperation(delta)
	ret, err := c.ExecuteTransaction(c.transactionCtx, op, true)
	if err != nil {
		return 0, log.OrtooError(err, "fail to execute operation")
	}
	return ret.(int32), nil
}

//ExecuteLocal is the
func (c *intCounter) ExecuteLocal(op interface{}) (interface{}, error) {
	iop := op.(*model.IncreaseOperation)
	//c.Logger.Info("delta:", proto.MarshalTextString(iop))
	return c.snapshot.increaseCommon(iop.Delta), nil
	//return nil, nil
}

func (c *intCounter) ExecuteRemote(op interface{}) (interface{}, error) {
	iop := op.(*model.IncreaseOperation)
	return c.snapshot.increaseCommon(iop.Delta), nil
}

func (c *intCounter) GetWired() datatypes.WiredDatatype {
	return c.WiredDatatypeImpl
}

func (c *intCounter) DoTransaction(tag string, transFunc func(intCounter IntCounterInTransaction) error) error {
	transactionCtx, err := c.BeginTransaction(tag, c.transactionCtx, true)
	defer c.EndTransaction(transactionCtx, true)
	clone := &intCounter{
		TransactionDatatypeImpl: c.TransactionDatatypeImpl,
		snapshot:                c.snapshot,
		transactionCtx:          transactionCtx,
	}
	err = transFunc(clone)
	if err != nil {
		c.SetTransactionFail()
		return log.OrtooError(err, "fail to do the transaction: '%s'", tag)
	}
	return nil
}

func (c *intCounter) GetSnapshot() datatypes.Snapshot {
	return c.snapshot
}

func (c *intCounter) SetSnapshot(snapshot datatypes.Snapshot) {
	c.snapshot = snapshot.(*intCounterSnapshot)
}

type intCounterSnapshot struct {
	value int32
}

func (c *intCounterSnapshot) CloneSnapshot() datatypes.Snapshot {
	return &intCounterSnapshot{
		value: c.value,
	}
}

func (c *intCounterSnapshot) increaseCommon(delta int32) int32 {
	temp := c.value
	c.value = c.value + delta
	log.Logger.Infof("increaseCommon: %d + %d = %d", temp, delta, c.value)
	return c.value
}
