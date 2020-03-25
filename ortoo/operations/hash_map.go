package operations

import (
	"github.com/knowhunger/ortoo/ortoo/model"
)

func NewPutOperation(key string, value interface{}) *PutOperation {
	return &PutOperation{
		BaseOperation: NewBaseOperation(nil),
		C: PutContent{
			Key:   key,
			Value: value,
		},
	}
}

type PutContent struct {
	Key   string
	Value interface{}
}

type PutOperation struct {
	*BaseOperation
	C PutContent
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
		PutContent
	}{
		ID:         its.BaseOperation.GetAsJSON(),
		Type:       model.TypeOfOperation_HASH_MAP_PUT.String(),
		PutContent: its.C,
	}
}

func NewRemoveOperation(key string) *RemoveOperation {
	return &RemoveOperation{
		BaseOperation: NewBaseOperation(nil),
		C: RemoveContent{
			Key: key,
		},
	}
}

type RemoveContent struct {
	Key string
}

type RemoveOperation struct {
	*BaseOperation
	C RemoveContent
}

func (its *RemoveOperation) ExecuteLocal(datatype model.Datatype) (interface{}, error) {
	return datatype.ExecuteLocal(its)
}

func (its *RemoveOperation) ExecuteRemote(datatype model.Datatype) (interface{}, error) {
	return datatype.ExecuteRemote(its)
}

func (its *RemoveOperation) ToModelOperation() *model.Operation {
	return &model.Operation{
		ID:     its.ID,
		OpType: model.TypeOfOperation_HASH_MAP_REMOVE,
		Json:   marshalContent(its.C),
	}
}

func (its *RemoveOperation) GetType() model.TypeOfOperation {
	return model.TypeOfOperation_HASH_MAP_REMOVE
}

func (its *RemoveOperation) String() string {
	return toString(its.ID, its.C)
}

func (its *RemoveOperation) GetAsJSON() interface{} {
	return &struct {
		ID   interface{}
		Type string
		RemoveContent
	}{
		ID:            its.BaseOperation.GetAsJSON(),
		Type:          model.TypeOfOperation_HASH_MAP_REMOVE.String(),
		RemoveContent: its.C,
	}
}
