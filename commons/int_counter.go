package commons

import (
	"encoding/json"
	"errors"
	"github.com/gogo/protobuf/types"
	"github.com/knowhunger/ortoo/commons/internal/datatypes"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
)

// IntCounter is an Ortoo datatype which provides int counter interfaces.
type IntCounter struct {
	datatypes.PublicWiredDatatypeInterface
	*datatypes.CommonDatatype
	Snapshot *IntCounterSnapshot
}

// IntCounterInTransaction is an Ortoo datatype which provides int counter interfaces in a transaction.
type IntCounterInTransaction interface {
	Get() int32
	Increase() (int32, error)
	IncreaseBy(delta int32) (int32, error)
}

// NewIntCounter creates a new int counter
func NewIntCounter(key string, cuid model.CUID, wire datatypes.Wire) (*IntCounter, error) {
	intCounter := &IntCounter{
		CommonDatatype: &datatypes.CommonDatatype{},
		Snapshot: &IntCounterSnapshot{
			Value: 0,
		},
	}
	err := intCounter.Initialize(key, model.TypeOfDatatype_INT_COUNTER, cuid, wire, intCounter.Snapshot, intCounter)
	if err != nil {
		return nil, log.OrtooErrorf(err, "fail to initialize intCounter")
	}
	return intCounter, nil
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
		CommonDatatype: &datatypes.CommonDatatype{
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

func (c *IntCounter) GetCommon() *datatypes.CommonDatatype {
	return c.CommonDatatype
}

// ExecuteLocal is the
func (c *IntCounter) ExecuteLocal(op interface{}) (interface{}, error) {
	iop := op.(*model.IncreaseOperation)
	// c.Logger.Info("delta:", proto.MarshalTextString(iop))
	return c.Snapshot.increaseCommon(iop.Delta), nil
	// return nil, nil
}

// ExecuteRemote is called by operation.ExecuteRemote()
func (c *IntCounter) ExecuteRemote(op interface{}) (interface{}, error) {
	switch o := op.(type) {
	case *model.SnapshotOperation:
		newSnap := IntCounterSnapshot{}
		if err := json.Unmarshal(o.Snapshot.Value, &newSnap); err != nil {
			return nil, log.OrtooError(err)
		}
		c.Snapshot = &newSnap
		return nil, nil
	case *model.IncreaseOperation:
		return c.Snapshot.increaseCommon(o.Delta), nil
	}

	return nil, log.OrtooError(errors.New("invalid operation"))
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

func (c *IntCounter) GetSnapshot() model.Snapshot {
	return c.Snapshot
}

func (c *IntCounter) SetSnapshot(snapshot model.Snapshot) {
	c.Snapshot = snapshot.(*IntCounterSnapshot)
}

func (c *IntCounter) GetMetaAndSnapshot() ([]byte, string, error) {
	meta, err := c.CommonDatatype.GetMeta()
	if err != nil {
		return nil, "", log.OrtooError(err)
	}
	jsonb, err := json.Marshal(c.Snapshot)
	if err != nil {
		return nil, "", log.OrtooError(err)
	}

	return meta, string(jsonb), nil
}

func (c *IntCounter) SetMetaAndSnapshot(meta []byte, snapshot string) error {
	if err := c.CommonDatatype.SetMeta(meta); err != nil {
		return log.OrtooError(err)
	}
	snap := &IntCounterSnapshot{}
	if err := json.Unmarshal([]byte(snapshot), snap); err != nil {
		return log.OrtooError(err)
	}
	c.Snapshot = snap
	return nil
}

//////////////////////////////////////////////////////////////////
//  IntCounterSnapshot
//////////////////////////////////////////////////////////////////

type IntCounterSnapshot struct {
	Value int32 `json:"value"`
}

func (i *IntCounterSnapshot) CloneSnapshot() model.Snapshot {
	return &IntCounterSnapshot{
		Value: i.Value,
	}
}

func (i *IntCounterSnapshot) GetTypeAny() (*types.Any, error) {
	bin, err := json.Marshal(i)
	if err != nil {
		return nil, log.OrtooError(err)
	}
	return &types.Any{
		TypeUrl: i.GetTypeUrl(),
		Value:   bin,
	}, nil
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
