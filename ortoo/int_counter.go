package ortoo

import (
	"encoding/json"
	operations2 "github.com/knowhunger/ortoo/ortoo/operations"

	// "errors"
	"fmt"
	"github.com/knowhunger/ortoo/ortoo/errors"
	"github.com/knowhunger/ortoo/ortoo/internal/datatypes"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/ortoo/model"
)

// IntCounter is an Ortoo datatype which provides int counter interfaces.
type IntCounter interface {
	Datatype
	IntCounterInTxn
	DoTransaction(tag string, txnFunc func(intCounter IntCounterInTxn) error) error
}

// IntCounterInTxn is an Ortoo datatype which provides int counter interfaces in a transaction.
type IntCounterInTxn interface {
	Get() int32
	Increase() (int32, error)
	IncreaseBy(delta int32) (int32, error)
}

type intCounter struct {
	*datatype
	snapshot *intCounterSnapshot
	// handler  *IntCounterHandlers
}

// newIntCounter creates a new int counter
func newIntCounter(key string, cuid model.CUID, wire datatypes.Wire, handler *Handlers) IntCounter {
	intCounter := &intCounter{
		datatype: &datatype{
			FinalDatatype: &datatypes.FinalDatatype{},
			handlers:      handler,
		},
		snapshot: &intCounterSnapshot{
			Value: 0,
		},
	}
	intCounter.Initialize(key, model.TypeOfDatatype_INT_COUNTER, cuid, wire, intCounter.snapshot, intCounter)
	return intCounter
}

func (its *intCounter) DoTransaction(tag string, txnFunc func(intCounter IntCounterInTxn) error) error {
	return its.FinalDatatype.DoTransaction(tag, func(txnCtx *datatypes.TransactionContext) error {
		clone := &intCounter{
			datatype: &datatype{
				FinalDatatype: &datatypes.FinalDatatype{
					TransactionDatatype: its.FinalDatatype.TransactionDatatype,
					TransactionCtx:      txnCtx,
				},
				handlers: its.handlers,
			},
			snapshot: its.snapshot,
		}
		return txnFunc(clone)
	})
}

func (its *intCounter) GetFinal() *datatypes.FinalDatatype {
	return its.FinalDatatype
}

// ExecuteLocal is the
func (its *intCounter) ExecuteLocal(op interface{}) (interface{}, error) {
	iop := op.(*operations2.IncreaseOperation)
	return its.snapshot.increaseCommon(iop.C.Delta), nil
}

// ExecuteRemote is called by operation.ExecuteRemote()
func (its *intCounter) ExecuteRemote(op interface{}) (interface{}, error) {
	switch cast := op.(type) {
	case *operations2.SnapshotOperation:
		newSnap := intCounterSnapshot{}
		if err := json.Unmarshal([]byte(cast.C.Snapshot), &newSnap); err != nil {
			return nil, errors.NewDatatypeError(errors.ErrDatatypeSnapshot, err.Error())
		}
		its.snapshot = &newSnap
		return nil, nil
	case *operations2.IncreaseOperation:
		return its.snapshot.increaseCommon(cast.C.Delta), nil
	}

	return nil, errors.NewDatatypeError(errors.ErrDatatypeIllegalOperation, op)
}

func (its *intCounter) Get() int32 {
	return its.snapshot.Value
}

func (its *intCounter) Increase() (int32, error) {
	return its.IncreaseBy(1)
}

func (its *intCounter) IncreaseBy(delta int32) (int32, error) {
	op := operations2.NewIncreaseOperation(delta)
	ret, err := its.ExecuteOperationWithTransaction(its.TransactionCtx, op, true)
	if err != nil {
		return 0, log.OrtooErrorf(err, "fail to execute operation")
	}
	return ret.(int32), nil
}

func (its *intCounter) GetSnapshot() model.Snapshot {
	return its.snapshot
}

func (its *intCounter) SetSnapshot(snapshot model.Snapshot) {
	its.snapshot = snapshot.(*intCounterSnapshot)
}

func (its *intCounter) GetAsJSON() (string, error) {
	j := &struct {
		Value int32
	}{
		Value: its.snapshot.Value,
	}
	b, err := json.Marshal(j)
	if err != nil {
		return "", errors.NewDatatypeError(errors.ErrDatatypeSnapshot, err.Error())
	}
	return string(b), nil
}

func (its *intCounter) GetMetaAndSnapshot() ([]byte, model.Snapshot, error) {
	meta, err := its.FinalDatatype.GetMeta()
	if err != nil {
		return nil, nil, errors.NewDatatypeError(errors.ErrDatatypeSnapshot, err.Error())
	}
	// jsonb, err := json.Marshal(its.snapshot)
	// if err != nil {
	// 	return nil, nil, errors.NewDatatypeError(errors.ErrDatatypeSnapshot, err.Error())
	// }

	return meta, its.snapshot, nil
}

func (its *intCounter) SetMetaAndSnapshot(meta []byte, snapshot model.Snapshot) error {
	if err := its.FinalDatatype.SetMeta(meta); err != nil {
		return errors.NewDatatypeError(errors.ErrDatatypeSnapshot, err.Error())
	}

	its.snapshot = snapshot.(*intCounterSnapshot)
	return nil
}

// ////////////////////////////////////////////////////////////////
//  intCounterSnapshot
// ////////////////////////////////////////////////////////////////

type intCounterSnapshot struct {
	Value int32 `json:"value"`
}

func (its *intCounterSnapshot) CloneSnapshot() model.Snapshot {
	return &intCounterSnapshot{
		Value: its.Value,
	}
}

func (its *intCounterSnapshot) GetAsJSON() (string, error) {
	j, err := json.Marshal(its)
	if err != nil {
		return "", errors.NewDatatypeError(errors.ErrDatatypeSnapshot, err.Error())
	}
	return string(j), nil
}

func (its *intCounterSnapshot) increaseCommon(delta int32) int32 {
	temp := its.Value
	its.Value = its.Value + delta
	log.Logger.Infof("increaseCommon: %d + %d = %d", temp, delta, its.Value)
	return its.Value
}

func (its *intCounterSnapshot) String() string {
	return fmt.Sprintf("Map: %d", its.Value)
}
