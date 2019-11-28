package commons

import (
	"github.com/knowhunger/ortoo/commons/internal/datatypes"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
)

// IntCounter is an Ortoo datatype which provides int counter interfaces.
type IntCounter interface {
	datatypes.PublicWiredDatatypeInterface
	datatypes.IntCounterInTransaction
	DoTransaction(tag string, transFunc func(intCounter datatypes.IntCounterInTransaction) error) error
}

// // IntCounterInTransaction is an Ortoo datatype which provides int counter interfaces in a transaction.
// type IntCounterInTransaction interface {
// 	Get() int32
// 	Increase() (int32, error)
// 	IncreaseBy(delta int32) (int32, error)
// }
//
// type intCounter struct {
// 	*datatypes.CommonDatatype
// 	snapshot *intCounterSnapshot
// }

// NewIntCounter creates a new int counter
func NewIntCounter(key string, cuid model.CUID, wire datatypes.Wire) (IntCounter, error) {
	snapshot := &datatypes.IntCounterSnapshot{
		Value: 0,
	}
	intCounter := &datatypes.IntCounter{
		CommonDatatype: &datatypes.CommonDatatype{},
		Snapshot:       snapshot,
	}
	err := intCounter.Initialize(key, model.TypeOfDatatype_INT_COUNTER, cuid, wire, snapshot, intCounter)
	if err != nil {
		return nil, log.OrtooErrorf(err, "fail to initialize intCounter")
	}
	return intCounter, nil
}

// func (c *intCounter) Get() int32 {
// 	return c.snapshot.Value
// }
//
// func (c *intCounter) Increase() (int32, error) {
// 	return c.IncreaseBy(1)
// }
//
// func (c *intCounter) IncreaseBy(delta int32) (int32, error) {
// 	op := model.NewIncreaseOperation(delta)
// 	ret, err := c.ExecuteOperationWithTransaction(c.TransactionCtx, op, true)
// 	if err != nil {
// 		return 0, log.OrtooErrorf(err, "fail to execute operation")
// 	}
// 	return ret.(int32), nil
// }
//
//
//
// func (c *intCounter) DoTransaction(tag string, transFunc func(intCounter IntCounterInTransaction) error) error {
// 	transactionCtx, err := c.BeginTransaction(tag, c.TransactionCtx, true)
// 	defer func() {
// 		if err := c.EndTransaction(transactionCtx, true, true); err != nil {
// 			_ = log.OrtooError(err)
// 		}
// 	}()
// 	// make a clone of intCounter have nil CommonDatatype.transactionCtx, which means
// 	clone := &intCounter{
// 		CommonDatatype: &datatypes.CommonDatatype{
// 			TransactionDatatype: c.CommonDatatype.TransactionDatatype,
// 			TransactionCtx:      transactionCtx,
// 		},
// 		snapshot: c.snapshot,
// 	}
// 	err = transFunc(clone)
// 	if err != nil {
// 		c.SetTransactionFail()
// 		return log.OrtooErrorf(err, "fail to do the transaction: '%s'", tag)
// 	}
// 	return nil
// }
//
// func (c *intCounter) GetSnapshot() model.Snapshot {
// 	return c.snapshot
// }
//
// func (c *intCounter) SetSnapshot(snapshot model.Snapshot) {
// 	c.snapshot = snapshot.(*intCounterSnapshot)
// }
//
// type intCounterSnapshot struct {
// 	Value int32 `json:"value"`
// }
//
// func (i *intCounterSnapshot) CloneSnapshot() model.Snapshot {
// 	return &intCounterSnapshot{
// 		Value: i.Value,
// 	}
// }
//
// func (i *intCounterSnapshot) GetTypeUrl() string {
// 	return "github.com/knowhunger/ortoo/common/intCounterSnapshot"
// }
//
// func (i *intCounterSnapshot) increaseCommon(delta int32) int32 {
// 	temp := i.Value
// 	i.Value = i.Value + delta
// 	log.Logger.Infof("increaseCommon: %d + %d = %d", temp, delta, i.Value)
// 	return i.Value
// }
