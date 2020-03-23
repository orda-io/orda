package operations

import (
	"github.com/knowhunger/ortoo/ortoo/model"
)

func NewIncreaseOperation(delta int32) *IncreaseOperation {
	return &IncreaseOperation{
		BaseOperation: NewBaseOperation(nil),
		C: IncreaseContent{
			Delta: delta,
		},
	}
}

type IncreaseContent struct {
	Delta int32
}

type IncreaseOperation struct {
	*BaseOperation
	C IncreaseContent
}

func (its *IncreaseOperation) ExecuteLocal(datatype model.Datatype) (interface{}, error) {
	return datatype.ExecuteLocal(its)
}

func (its *IncreaseOperation) ExecuteRemote(datatype model.Datatype) (interface{}, error) {
	return datatype.ExecuteRemote(its)
}

func (its *IncreaseOperation) ToModelOperation() *model.Operation {
	return &model.Operation{
		ID:     its.ID,
		OpType: model.TypeOfOperation_INT_COUNTER_INCREASE,
		Json:   marshalContent(its.C),
	}
}

func (its *IncreaseOperation) GetType() model.TypeOfOperation {
	return model.TypeOfOperation_INT_COUNTER_INCREASE
}

func (its *IncreaseOperation) String() string {
	return toString(its.ID, its.C)
}

func (its *IncreaseOperation) GetAsJSON() interface{} {
	return &struct {
		ID   interface{}
		Type string
		IncreaseContent
	}{
		ID:              its.BaseOperation.GetAsJSON(),
		Type:            model.TypeOfOperation_INT_COUNTER_INCREASE.String(),
		IncreaseContent: its.C,
	}
}
