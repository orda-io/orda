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
	DoTransaction(tag string, txFunc func(counter CounterInTxn) error) error
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
func newCounter(base *datatypes.BaseDatatype, wire iface.Wire, handler *Handlers) (Counter, errors.OrtooError) {
	counter := &counter{
		datatype: &datatype{
			ManageableDatatype: &datatypes.ManageableDatatype{},
			handlers:           handler,
		},
		SnapshotDatatype: &datatypes.SnapshotDatatype{
			Snapshot: newCounterSnapshot(base),
		},
	}
	return counter, counter.Initialize(base, wire, counter.GetSnapshot(), counter)
}

func (its *counter) DoTransaction(tag string, txFunc func(counter CounterInTxn) error) error {
	return its.ManageableDatatype.DoTransaction(tag, func(txCtx *datatypes.TransactionContext) error {
		clone := &counter{
			datatype:         its.newDatatype(txCtx),
			SnapshotDatatype: its.SnapshotDatatype,
		}
		return txFunc(clone)
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

	return nil, errors.DatatypeIllegalParameters.New(its.L(), op)
}

func (its *counter) ResetSnapshot() {
	its.SnapshotDatatype.SetSnapshot(newCounterSnapshot(its.BaseDatatype))
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
	ret, err := its.SentenceInTransaction(its.TransactionCtx, op, true)
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

func (its *counterSnapshot) GetAsJSONCompatible() interface{} {
	return its
}

func (its *counterSnapshot) increaseCommon(delta int32) int32 {
	its.Value += delta
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
