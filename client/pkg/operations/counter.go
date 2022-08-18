package operations

import (
	"github.com/orda-io/orda/client/pkg/model"
)

type increaseBody struct {
	Delta int32
}

// NewIncreaseOperation creates an IncreaseOperation.
func NewIncreaseOperation(delta int32) *IncreaseOperation {
	return &IncreaseOperation{
		baseOperation: newBaseOperation(
			model.TypeOfOperation_COUNTER_INCREASE,
			nil,
			&increaseBody{
				Delta: delta,
			},
		),
	}
}

// IncreaseOperation is used to increase value to IntCounter.
type IncreaseOperation struct {
	baseOperation
}

// GetBody returns the body
func (its *IncreaseOperation) GetBody() int32 {
	return its.Body.(*increaseBody).Delta
}
