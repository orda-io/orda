package operations

import (
	"github.com/knowhunger/ortoo/pkg/model"
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

func (its *IncreaseOperation) GetBody() *increaseBody {
	return its.Body.(*increaseBody)
}
