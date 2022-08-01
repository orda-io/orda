package orda

import (
	"encoding/json"
	"fmt"
	"github.com/orda-io/orda/client/pkg/errors"
	"github.com/orda-io/orda/client/pkg/iface"
	"github.com/orda-io/orda/client/pkg/internal/datatypes"
	"github.com/orda-io/orda/client/pkg/operations"
)

// Counter is an Orda datatype which provides int counter interfaces.
type Counter interface {
	Datatype
	CounterInTx
	Transaction(tag string, txFunc func(counter CounterInTx) error) error
}

// CounterInTx is an Orda datatype which provides int counter interfaces in a transaction.
type CounterInTx interface {
	Get() int32
	Increase() (int32, errors.OrdaError)
	IncreaseBy(delta int32) (int32, errors.OrdaError)
}

type counter struct {
	*datatype
	*datatypes.SnapshotDatatype
}

// newCounter creates a new counter
func newCounter(base *datatypes.BaseDatatype, wire iface.Wire, handlers *Handlers) (Counter, errors.OrdaError) {
	counter := &counter{
		datatype:         newDatatype(base, wire, handlers),
		SnapshotDatatype: datatypes.NewSnapshotDatatype(base, nil),
	}
	return counter, counter.init(counter)
}

func (its *counter) Transaction(tag string, txFunc func(counter CounterInTx) error) error {
	return its.DoTransaction(tag, its.TxCtx, func(txCtx *datatypes.TransactionContext) error {
		clone := &counter{
			datatype:         its.cloneDatatype(txCtx),
			SnapshotDatatype: its.SnapshotDatatype,
		}
		return txFunc(clone)
	})
}

// ExecuteLocal enables the operation to perform something at the local client.
func (its *counter) ExecuteLocal(op interface{}) (interface{}, errors.OrdaError) {
	switch cast := op.(type) {
	case *operations.IncreaseOperation:
		return its.snapshot().increaseCommon(cast.GetBody().Delta), nil
	}
	return nil, errors.DatatypeIllegalOperation.New(its.L(), its.TypeOf.String(), op)
}

// ExecuteRemote is called by operation.ExecuteRemote()
func (its *counter) ExecuteRemote(op interface{}) (interface{}, errors.OrdaError) {
	switch cast := op.(type) {
	case *operations.SnapshotOperation:
		return nil, its.ApplySnapshot(cast.GetBody())
	case *operations.IncreaseOperation:
		return its.snapshot().increaseCommon(cast.GetBody().Delta), nil
	}

	return nil, errors.DatatypeIllegalOperation.New(its.L(), its.TypeOf.String(), op)
}

func (its *counter) ResetSnapshot() {
	its.Snapshot = newCounterSnapshot(its.BaseDatatype)
}

func (its *counter) Get() int32 {
	return its.snapshot().Value
}

func (its *counter) Increase() (int32, errors.OrdaError) {
	return its.IncreaseBy(1)
}

func (its *counter) snapshot() *counterSnapshot {
	return its.GetSnapshot().(*counterSnapshot)
}

func (its *counter) IncreaseBy(delta int32) (int32, errors.OrdaError) {
	op := operations.NewIncreaseOperation(delta)
	ret, err := its.SentenceInTx(its.TxCtx, op, true)
	if err != nil {
		return its.snapshot().Value, err
	}
	return ret.(int32), nil
}

func (its *counter) ToJSON() interface{} {
	return struct {
		Counter interface{}
	}{
		Counter: its.snapshot().ToJSON(),
	}
}

// ////////////////////////////////////////////////////////////////
//  counterSnapshot
// ////////////////////////////////////////////////////////////////

type counterSnapshot struct {
	iface.BaseDatatype
	Value int32
}

func newCounterSnapshot(base iface.BaseDatatype) *counterSnapshot {
	return &counterSnapshot{
		BaseDatatype: base,
		Value:        0,
	}
}

func (its *counterSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct{ Counter int32 }{Counter: its.Value})
}

func (its *counterSnapshot) UnmarshalJSON(bytes []byte) error {
	var unmarshal struct{ Counter int32 }
	if err := json.Unmarshal(bytes, &unmarshal); err != nil {
		return err
	}
	its.Value = unmarshal.Counter
	return nil
}

func (its *counterSnapshot) increaseCommon(delta int32) int32 {
	its.Value += delta
	return its.Value
}

func (its *counterSnapshot) String() string {
	return fmt.Sprintf("Counter: %d", its.Value)
}

func (its *counterSnapshot) ToJSON() interface{} {
	return its.Value
}
