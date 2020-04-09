package operations

import (
	"github.com/knowhunger/ortoo/ortoo/iface"
	"github.com/knowhunger/ortoo/ortoo/model"
)

// NewIncreaseOperation creates an IncreaseOperation.
func NewIncreaseOperation(delta int32) *IncreaseOperation {
	return &IncreaseOperation{
		baseOperation: newBaseOperation(nil),
		C: increaseContent{
			Delta: delta,
		},
	}
}

type increaseContent struct {
	Delta int32
}

// IncreaseOperation is used to increase value to IntCounter.
type IncreaseOperation struct {
	*baseOperation
	C increaseContent
}

// ExecuteLocal enables the operation to perform something at the local client.
func (its *IncreaseOperation) ExecuteLocal(datatype iface.Datatype) (interface{}, error) {
	return datatype.ExecuteLocal(its)
}

// ExecuteRemote enables the operation to perform something at the remote clients.
func (its *IncreaseOperation) ExecuteRemote(datatype iface.Datatype) (interface{}, error) {
	return datatype.ExecuteRemote(its)
}

// ToModelOperation transforms this operation to the model.Operation.
func (its *IncreaseOperation) ToModelOperation() *model.Operation {
	return &model.Operation{
		ID:     its.ID,
		OpType: model.TypeOfOperation_INT_COUNTER_INCREASE,
		Json:   marshalContent(its.C),
	}
}

// GetType returns the type of operation.
func (its *IncreaseOperation) GetType() model.TypeOfOperation {
	return model.TypeOfOperation_INT_COUNTER_INCREASE
}

func (its *IncreaseOperation) String() string {
	return toString(its.ID, its.C)
}

// GetAsJSON returns the operation in the format of JSON compatible struct.
func (its *IncreaseOperation) GetAsJSON() interface{} {
	return struct {
		ID   interface{}
		Type string
		increaseContent
	}{
		ID:              its.baseOperation.GetAsJSON(),
		Type:            model.TypeOfOperation_INT_COUNTER_INCREASE.String(),
		increaseContent: its.C,
	}
}
