package commons

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gogo/protobuf/types"
	"github.com/knowhunger/ortoo/commons/internal/datatypes"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
)

// IntCounter is an Ortoo datatype which provides int counter interfaces.
type IntCounter interface {
	datatypes.PublicWiredDatatypeInterface
	IntCounterInTxn
	DoTransaction(tag string, transFunc func(intCounter IntCounterInTxn) error) error
}

// IntCounterInTxn is an Ortoo datatype which provides int counter interfaces in a transaction.
type IntCounterInTxn interface {
	Get() int32
	Increase() (int32, error)
	IncreaseBy(delta int32) (int32, error)
}

type intCounter struct {
	*datatypes.CommonDatatype
	snapshot *intCounterSnapshot
	handler  *IntCounterHandlers
}

// NewIntCounter creates a new int counter
func NewIntCounter(key string, cuid model.CUID, wire datatypes.Wire, handler *IntCounterHandlers) (IntCounter, error) {

	intCounter := &intCounter{
		CommonDatatype: &datatypes.CommonDatatype{},
		snapshot: &intCounterSnapshot{
			Value: 0,
		},
		handler: handler,
	}
	if err := intCounter.Initialize(key, model.TypeOfDatatype_INT_COUNTER, cuid, wire, intCounter.snapshot, intCounter); err != nil {
		return nil, log.OrtooErrorf(err, "fail to initialize intCounter")
	}
	return intCounter, nil
}

func (c *intCounter) DoTransaction(tag string, transFunc func(intCounter IntCounterInTxn) error) error {
	transactionCtx, err := c.BeginTransaction(tag, c.TransactionCtx, true)
	defer func() {
		if err := c.EndTransaction(transactionCtx, true, true); err != nil {
			_ = log.OrtooError(err)
		}
	}()
	// make a clone of intCounter have nil CommonDatatype.transactionCtx, which means
	clone := &intCounter{
		CommonDatatype: &datatypes.CommonDatatype{
			TransactionDatatype: c.CommonDatatype.TransactionDatatype,
			TransactionCtx:      transactionCtx,
		},
		snapshot: c.snapshot,
	}
	err = transFunc(clone)
	if err != nil {
		c.SetTransactionFail()
		return log.OrtooErrorf(err, "fail to do the transaction: '%s'", tag)
	}
	return nil
}

func (c *intCounter) GetCommon() *datatypes.CommonDatatype {
	return c.CommonDatatype
}

// ExecuteLocal is the
func (c *intCounter) ExecuteLocal(op interface{}) (interface{}, error) {
	iop := op.(*model.IncreaseOperation)
	return c.snapshot.increaseCommon(iop.Delta), nil
}

// ExecuteRemote is called by operation.ExecuteRemote()
func (c *intCounter) ExecuteRemote(op interface{}) (interface{}, error) {
	switch o := op.(type) {
	case *model.SnapshotOperation:
		newSnap := intCounterSnapshot{}
		if err := json.Unmarshal(o.Snapshot.Value, &newSnap); err != nil {
			return nil, log.OrtooError(err)
		}
		c.snapshot = &newSnap
		return nil, nil
	case *model.IncreaseOperation:
		return c.snapshot.increaseCommon(o.Delta), nil
	}

	return nil, log.OrtooError(errors.New("invalid operation"))
}

func (c *intCounter) HandleSubscription() {
	if c.handler != nil && c.handler.subscribeHandler != nil {
		go c.handler.subscribeHandler(c)
	}

}

func (c *intCounter) HandleError(err error) {
	if c.handler != nil && c.handler.errorHandler != nil {
		go c.handler.errorHandler(err)
	}
}

func (c *intCounter) HandleRemoteChange() {

}

func (c *intCounter) Get() int32 {
	return c.snapshot.Value
}

func (c *intCounter) Increase() (int32, error) {
	return c.IncreaseBy(1)
}

func (c *intCounter) IncreaseBy(delta int32) (int32, error) {
	op := model.NewIncreaseOperation(delta)
	ret, err := c.ExecuteOperationWithTransaction(c.TransactionCtx, op, true)
	if err != nil {
		return 0, log.OrtooErrorf(err, "fail to execute operation")
	}
	return ret.(int32), nil
}

func (c *intCounter) GetSnapshot() model.Snapshot {
	return c.snapshot
}

func (c *intCounter) SetSnapshot(snapshot model.Snapshot) {
	c.snapshot = snapshot.(*intCounterSnapshot)
}

func (c *intCounter) GetMetaAndSnapshot() ([]byte, string, error) {
	meta, err := c.CommonDatatype.GetMeta()
	if err != nil {
		return nil, "", log.OrtooError(err)
	}
	jsonb, err := json.Marshal(c.snapshot)
	if err != nil {
		return nil, "", log.OrtooError(err)
	}

	return meta, string(jsonb), nil
}

func (c *intCounter) SetMetaAndSnapshot(meta []byte, snapshot string) error {
	if err := c.CommonDatatype.SetMeta(meta); err != nil {
		return log.OrtooError(err)
	}
	snap := &intCounterSnapshot{}
	if err := json.Unmarshal([]byte(snapshot), snap); err != nil {
		return log.OrtooError(err)
	}
	c.snapshot = snap
	return nil
}

type IntCounterHandlers struct {
	subscribeHandler    func(intCounter IntCounter)
	remoteChangeHandler func(intCount IntCounter, opList []model.Operation)
	errorHandler        func(err error)
}

func NewIntCounterHandlers(
	subscribeHandler func(intCounter IntCounter),
	remoteChangeHandler func(intCounter IntCounter, opList []model.Operation),
	errorHandler func(err error)) *IntCounterHandlers {
	return &IntCounterHandlers{
		subscribeHandler:    subscribeHandler,
		remoteChangeHandler: remoteChangeHandler,
		errorHandler:        errorHandler,
	}
}

func (i *IntCounterHandlers) SetHandlers(
	subscribeHandler func(intCounter IntCounter),
	remoteChangeHandler func(intCounter IntCounter, opList []model.Operation),
	errHandler func(err error)) {
	i.subscribeHandler = subscribeHandler
	i.errorHandler = errHandler
	i.remoteChangeHandler = remoteChangeHandler
}

// ////////////////////////////////////////////////////////////////
//  intCounterSnapshot
// ////////////////////////////////////////////////////////////////

type intCounterSnapshot struct {
	Value int32 `json:"value"`
}

func (i *intCounterSnapshot) CloneSnapshot() model.Snapshot {
	return &intCounterSnapshot{
		Value: i.Value,
	}
}

func (i *intCounterSnapshot) GetTypeAny() (*types.Any, error) {
	bin, err := json.Marshal(i)
	if err != nil {
		return nil, log.OrtooError(err)
	}
	return &types.Any{
		TypeUrl: i.GetTypeUrl(),
		Value:   bin,
	}, nil
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

func (i *intCounterSnapshot) String() string {
	return fmt.Sprintf("Value: %d", i.Value)
}
