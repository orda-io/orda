package datatypes

import (
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
)

// IntCounterInTransaction is an Ortoo datatype which provides int counter interfaces in a transaction.
type IntCounterInTransaction interface {
	Get() int32
	Increase() (int32, error)
	IncreaseBy(delta int32) (int32, error)
}

type IntCounter struct {
	*CommonDatatype
	Snapshot *IntCounterSnapshot
}

// ExecuteLocal is the
func (c *IntCounter) ExecuteLocal(op interface{}) (interface{}, error) {
	iop := op.(*model.IncreaseOperation)
	// c.Logger.Info("delta:", proto.MarshalTextString(iop))
	return c.Snapshot.increaseCommon(iop.Delta), nil
	// return nil, nil
}

func (c *IntCounter) ExecuteRemote(op interface{}) (interface{}, error) {
	iop := op.(*model.IncreaseOperation)
	return c.Snapshot.increaseCommon(iop.Delta), nil
}

func (c *IntCounter) Get() int32 {
	return c.Snapshot.Value
}

func (c *IntCounter) Increase() (int32, error) {
	return c.IncreaseBy(1)
}

func (c *IntCounter) IncreaseBy(delta int32) (int32, error) {
	op := model.NewIncreaseOperation(delta)
	ret, err := c.ExecuteOperationWithTransaction(c.TransactionCtx, op, true)
	if err != nil {
		return 0, log.OrtooErrorf(err, "fail to execute operation")
	}
	return ret.(int32), nil
}

func (c *IntCounter) DoTransaction(tag string, transFunc func(intCounter IntCounterInTransaction) error) error {
	transactionCtx, err := c.BeginTransaction(tag, c.TransactionCtx, true)
	defer func() {
		if err := c.EndTransaction(transactionCtx, true, true); err != nil {
			_ = log.OrtooError(err)
		}
	}()
	// make a clone of intCounter have nil CommonDatatype.transactionCtx, which means
	clone := &IntCounter{
		CommonDatatype: &CommonDatatype{
			TransactionDatatype: c.CommonDatatype.TransactionDatatype,
			TransactionCtx:      transactionCtx,
		},
		Snapshot: c.Snapshot,
	}
	err = transFunc(clone)
	if err != nil {
		c.SetTransactionFail()
		return log.OrtooErrorf(err, "fail to do the transaction: '%s'", tag)
	}
	return nil
}

func (c *IntCounter) GetCommon() *CommonDatatype {
	return c.CommonDatatype
}

func (c *IntCounter) GetSnapshot() model.Snapshot {
	return c.Snapshot
}

func (c *IntCounter) SetSnapshot(snapshot model.Snapshot) {
	c.Snapshot = snapshot.(*IntCounterSnapshot)
}

func (c *IntCounter) GetMetaAndSnapshot() ([]byte, string) {

	return nil, nil
}

type IntCounterSnapshot struct {
	Value int32 `json:"value"`
}

func (i *IntCounterSnapshot) CloneSnapshot() model.Snapshot {
	return &IntCounterSnapshot{
		Value: i.Value,
	}
}

func (i *IntCounterSnapshot) GetTypeUrl() string {
	return "github.com/knowhunger/ortoo/common/intCounterSnapshot"
}

func (i *IntCounterSnapshot) increaseCommon(delta int32) int32 {
	temp := i.Value
	i.Value = i.Value + delta
	log.Logger.Infof("increaseCommon: %d + %d = %d", temp, delta, i.Value)
	return i.Value
}
