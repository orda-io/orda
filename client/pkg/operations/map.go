package operations

import (
	"github.com/orda-io/orda/client/pkg/model"
)

// PutBody is the body of PutOperation
type PutBody struct {
	Key   string
	Value interface{}
}

// NewPutOperation creates a PutOperation of hash map.
func NewPutOperation(key string, value interface{}) *PutOperation {
	return &PutOperation{
		baseOperation: newBaseOperation(
			model.TypeOfOperation_MAP_PUT,
			nil,
			&PutBody{
				Key:   key,
				Value: value,
			},
		),
	}
}

// PutOperation is used to put something in the hash map.
type PutOperation struct {
	baseOperation
}

// GetBody returns the body
func (its *PutOperation) GetBody() *PutBody {
	return its.Body.(*PutBody)
}

// ////////////////// RemoveOperation ////////////////////

// RemoveBody is the body of RemoveOperation
type RemoveBody struct {
	Key string
}

// NewRemoveOperation creates a RemoveOperation of hash map.
func NewRemoveOperation(key string) *RemoveOperation {
	return &RemoveOperation{
		baseOperation: newBaseOperation(
			model.TypeOfOperation_MAP_REMOVE,
			nil,
			&RemoveBody{
				Key: key,
			},
		),
	}
}

// RemoveOperation is used to remove something in the hash map.
type RemoveOperation struct {
	baseOperation
}

// GetBody returns the body
func (its *RemoveOperation) GetBody() *RemoveBody {
	return its.Body.(*RemoveBody)
}
