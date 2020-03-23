package operations

import (
	"github.com/knowhunger/ortoo/ortoo/model"
)

func NewPutOperation(key string, value interface{}) *PutOperation {
	return &PutOperation{
		BaseOperation: NewBaseOperation(nil),
		C: putContent{
			Key:   key,
			Value: value,
		},
	}
}

type putContent struct {
	Key   string
	Value interface{}
}

type PutOperation struct {
	*BaseOperation
	C putContent
}

func (its *PutOperation) ExecuteLocal(datatype model.Datatype) (interface{}, error) {
	return datatype.ExecuteLocal(its)
}

func (its *PutOperation) ExecuteRemote(datatype model.Datatype) (interface{}, error) {
	return datatype.ExecuteRemote(its)
}

func (its *PutOperation) ToModelOperation() *model.Operation {
	return &model.Operation{
		ID:     its.ID,
		OpType: model.TypeOfOperation_HASH_MAP_PUT,
		Json:   marshalContent(its.C),
	}
}

func (its *PutOperation) GetType() model.TypeOfOperation {
	return model.TypeOfOperation_HASH_MAP_PUT
}

func (its *PutOperation) String() string {
	return toString(its.ID, its.C)
}

func (its *PutOperation) GetAsJSON() interface{} {
	return &struct {
		ID   interface{}
		Type string
		putContent
	}{
		ID:         its.BaseOperation.GetAsJSON(),
		Type:       model.TypeOfOperation_HASH_MAP_PUT.String(),
		putContent: its.C,
	}
}
