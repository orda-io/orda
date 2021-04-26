package operations

import (
	"github.com/knowhunger/ortoo/pkg/model"
)

// NewPutOperation creates a PutOperation of hash map.
func NewPutOperation(key string, value interface{}) *PutOperation {
	return &PutOperation{
		baseOperation: newBaseOperation(nil),
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

// PutOperation is used to put something in the hash map.
type PutOperation struct {
	*baseOperation
	C putContent
}

// ToModelOperation transforms this operation to the model.Operation.
func (its *PutOperation) ToModelOperation() *model.Operation {
	return &model.Operation{
		ID:     its.ID,
		OpType: model.TypeOfOperation_MAP_PUT,
		Body:   marshalContent(its.C),
	}
}

// GetType returns the type of operation.
func (its *PutOperation) GetType() model.TypeOfOperation {
	return model.TypeOfOperation_MAP_PUT
}

func (its *PutOperation) String() string {
	return its.toString(its.GetType(), its.C)
}

// GetAsJSON returns the operation in the format of JSON compatible struct.
func (its *PutOperation) GetAsJSON() interface{} {
	return struct {
		ID   interface{}
		Type string
		putContent
	}{
		ID:         its.baseOperation.GetAsJSON(),
		Type:       model.TypeOfOperation_MAP_PUT.String(),
		putContent: its.C,
	}
}

// ////////////////// RemoveOperation ////////////////////

// NewRemoveOperation creates a RemoveOperation of hash map.
func NewRemoveOperation(key string) *RemoveOperation {
	return &RemoveOperation{
		baseOperation: newBaseOperation(nil),
		C: removeContent{
			Key: key,
		},
	}
}

type removeContent struct {
	Key string
}

// RemoveOperation is used to remove something in the hash map.
type RemoveOperation struct {
	*baseOperation
	C removeContent
}

// ToModelOperation transforms this operation to the model.Operation.
func (its *RemoveOperation) ToModelOperation() *model.Operation {
	return &model.Operation{
		ID:     its.ID,
		OpType: model.TypeOfOperation_MAP_REMOVE,
		Body:   marshalContent(its.C),
	}
}

// GetType returns the type of operation.
func (its *RemoveOperation) GetType() model.TypeOfOperation {
	return model.TypeOfOperation_MAP_REMOVE
}

func (its *RemoveOperation) String() string {
	return its.toString(its.GetType(), its.C)
}

// GetAsJSON returns the operation in the format of JSON compatible struct.
func (its *RemoveOperation) GetAsJSON() interface{} {
	return struct {
		ID   interface{}
		Type string
		removeContent
	}{
		ID:            its.baseOperation.GetAsJSON(),
		Type:          model.TypeOfOperation_MAP_REMOVE.String(),
		removeContent: its.C,
	}
}
