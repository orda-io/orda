package ortoo

import (
	"fmt"
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/iface"
	"github.com/knowhunger/ortoo/pkg/internal/datatypes"
	"github.com/knowhunger/ortoo/pkg/operations"
)

// Counter is an Ortoo datatype which provides int counter interfaces.
type Counter interface {
	Datatype
	CounterInTxn
	DoTransaction(tag string, txnFunc func(intCounter CounterInTxn) error) error
}

// CounterInTxn is an Ortoo datatype which provides int counter interfaces in a transaction.
type CounterInTxn interface {
	Get() int32
	Increase() (int32, errors.OrtooError)
	IncreaseBy(delta int32) (int32, errors.OrtooError)
}

type counter struct {
	*datatype
	*datatypes.SnapshotDatatype
}

// newCounter creates a new int counter
func newCounter(base *datatypes.BaseDatatype, wire iface.Wire, handler *Handlers) Counter {
	counter := &counter{
		datatype: &datatype{
			ManageableDatatype: &datatypes.ManageableDatatype{},
			handlers:           handler,
		},
		SnapshotDatatype: &datatypes.SnapshotDatatype{
			Snapshot: newCounterSnapshot(base),
		},
	}
	counter.Initialize(base, wire, counter.GetSnapshot(), counter)
	return counter
}

func (its *counter) DoTransaction(tag string, txnFunc func(intCounter CounterInTxn) error) error {
	return its.ManageableDatatype.DoTransaction(tag, func(txnCtx *datatypes.TransactionContext) error {
		clone := &counter{
			datatype: &datatype{
				ManageableDatatype: &datatypes.ManageableDatatype{
					TransactionDatatype: its.ManageableDatatype.TransactionDatatype,
					TransactionCtx:      txnCtx,
				},
				handlers: its.handlers,
			},
			SnapshotDatatype: its.SnapshotDatatype,
		}
		return txnFunc(clone)
	})
}

// ExecuteLocal enables the operation to perform something at the local client.
func (its *counter) ExecuteLocal(op interface{}) (interface{}, errors.OrtooError) {
	iop := op.(*operations.IncreaseOperation)
	return its.snapshot().increaseCommon(iop.C.Delta), nil
}

// ExecuteRemote is called by operation.ExecuteRemote()
func (its *counter) ExecuteRemote(op interface{}) (interface{}, errors.OrtooError) {
	switch cast := op.(type) {
	case *operations.SnapshotOperation:
		err := its.ApplySnapshotOperation(cast.GetContent(), newCounterSnapshot(its.BaseDatatype))
		return nil, err
	case *operations.IncreaseOperation:
		return its.snapshot().increaseCommon(cast.C.Delta), nil
	}

	return nil, errors.DatatypeIllegalParameters.New(its.Logger, op)
}

func (its *counter) Get() int32 {
	return its.snapshot().Value
}

func (its *counter) Increase() (int32, errors.OrtooError) {
	return its.IncreaseBy(1)
}

func (its *counter) snapshot() *counterSnapshot {
	return its.GetSnapshot().(*counterSnapshot)
}

func (its *counter) IncreaseBy(delta int32) (int32, errors.OrtooError) {
	op := operations.NewIncreaseOperation(delta)
	ret, err := its.ExecuteOperationWithTransaction(its.TransactionCtx, op, true)
	if err != nil {
		return its.snapshot().Value, err
	}
	return ret.(int32), nil
}

// ////////////////////////////////////////////////////////////////
//  counterSnapshot
// ////////////////////////////////////////////////////////////////

type counterSnapshot struct {
	base
	Value int32 `json:"value"`
}

func newCounterSnapshot(base iface.BaseDatatype) *counterSnapshot {
	return &counterSnapshot{
		base:  base,
		Value: 0,
	}
}

func (its *counterSnapshot) CloneSnapshot() iface.Snapshot {
	return &counterSnapshot{
		base:  its.base,
		Value: its.Value,
	}
}

func (its *counterSnapshot) GetAsJSONCompatible() interface{} {
	return its
}

func (its *counterSnapshot) increaseCommon(delta int32) int32 {
	its.Value = its.Value + delta
	return its.Value
}

func (its *counterSnapshot) String() string {
	return fmt.Sprintf("Map: %d", its.Value)
}

func (its *counterSnapshot) GetBase() iface.BaseDatatype {
	return its.base
}

func (its *counterSnapshot) SetBase(base iface.BaseDatatype) {
	its.base = base
}
