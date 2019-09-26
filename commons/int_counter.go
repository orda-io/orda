package commons

import (
	"github.com/knowhunger/ortoo/commons/internal/datatypes"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
)

//IntCounter is an Ortoo datatype which provides int counter interfaces.
type IntCounter interface {
	datatypes.PublicWiredDatatypeInterface
	IntCounterInTransaction
	DoTransaction(tag string, transFunc func(intCounter IntCounterInTransaction) error) error
}

//IntCounterInTransaction is an Ortoo datatype which provides int counter interfaces in a transaction.
type IntCounterInTransaction interface {
	Get() int32
	Increase() (int32, error)
	IncreaseBy(delta int32) (int32, error)
}

type intCounter struct {
	*datatypes.CommonDatatype
	snapshot *intCounterSnapshot
}

//NewIntCounter creates a new int counter
func NewIntCounter(key string, client Client) (IntCounter, error) {
	snapshot := &intCounterSnapshot{
		Value: 0,
	}
	intCounter := &intCounter{
		CommonDatatype: &datatypes.CommonDatatype{},
		snapshot:       snapshot,
	}
	ci := client.(*clientImpl)
	err := intCounter.Initialize(
		key,
		model.TypeOfDatatype_INT_COUNTER,
		ci.model.Cuid,
		ci.dataMgr,
		snapshot,
		intCounter)
	if err != nil {
		return nil, log.OrtooError(err, "fail to initialize intCounter")
	}
	return intCounter, nil
}

func (c *intCounter) Get() int32 {
	return c.snapshot.Value
}

func (c *intCounter) Increase() (int32, error) {
	return c.IncreaseBy(1)
}

func (c *intCounter) IncreaseBy(delta int32) (int32, error) {
	op := model.NewIncreaseOperation(delta)
	ret, err := c.ExecuteTransaction(c.TransactionCtx, op, true)
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
	transactionCtx, err := c.BeginTransaction(tag, c.TransactionCtx, true)
	defer c.EndTransaction(transactionCtx, true)
	clone := &intCounter{
		CommonDatatype: c.CommonDatatype,
		snapshot:       c.snapshot,
	}
	err = transFunc(clone)
	if err != nil {
		c.SetTransactionFail()
		return log.OrtooError(err, "fail to do the transaction: '%s'", tag)
	}
	return nil
}

func (c *intCounter) GetSnapshot() model.Snapshot {
	return c.snapshot
}

func (c *intCounter) SetSnapshot(snapshot model.Snapshot) {
	c.snapshot = snapshot.(*intCounterSnapshot)
}

type intCounterSnapshot struct {
	Value int32 `json:"value"`
}

func (i *intCounterSnapshot) CloneSnapshot() model.Snapshot {
	return &intCounterSnapshot{
		Value: i.Value,
	}
}

func (i *intCounterSnapshot) GetTypeUrl() string {
	return "github.com/knowhunger/ortoo/common/intCounterSnapshot"
}

func (i *intCounterSnapshot) increaseCommon(delta int32) int32 {
	temp := i.Value
	i.Value = i.Value + delta
	log.Logger.Infof("increaseCommon: %d + %d = %d", temp, delta, i.Value)
	return i.Value
}
