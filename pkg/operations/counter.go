package operations

import (
	"github.com/knowhunger/ortoo/pkg/model"
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

// ToModelOperation transforms this operation to the model.Operation.
func (its *IncreaseOperation) ToModelOperation() *model.Operation {
	return &model.Operation{
		ID:     its.ID,
		OpType: model.TypeOfOperation_COUNTER_INCREASE,
		Body:   marshalContent(its.C),
	}
}

// GetType returns the type of operation.
func (its *IncreaseOperation) GetType() model.TypeOfOperation {
	return model.TypeOfOperation_COUNTER_INCREASE
}

func (its *IncreaseOperation) String() string {
	return its.toString(its.GetType(), its.C)
}

// GetAsJSON returns the operation in the format of JSON compatible struct.
func (its *IncreaseOperation) GetAsJSON() interface{} {
	return struct {
		ID   interface{}
		Type string
		increaseContent
	}{
		ID:              its.baseOperation.GetAsJSON(),
		Type:            model.TypeOfOperation_COUNTER_INCREASE.String(),
		increaseContent: its.C,
	}
}
