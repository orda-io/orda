package operations

import (
	"github.com/orda-io/orda/client/pkg/model"
)

type putBody struct {
	Key   string
	Value interface{}
}

// NewPutOperation creates a PutOperation of hash map.
func NewPutOperation(key string, value interface{}) *PutOperation {
	return &PutOperation{
		baseOperation: newBaseOperation(
			model.TypeOfOperation_MAP_PUT,
			nil,
			&putBody{
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

func (its *PutOperation) GetBody() *putBody {
	return its.Body.(*putBody)
}

// ////////////////// RemoveOperation ////////////////////

type removeBody struct {
	Key string
}

// NewRemoveOperation creates a RemoveOperation of hash map.
func NewRemoveOperation(key string) *RemoveOperation {
	return &RemoveOperation{
		baseOperation: newBaseOperation(
			model.TypeOfOperation_MAP_REMOVE,
			nil,
			&removeBody{
				Key: key,
			},
		),
	}
}

// RemoveOperation is used to remove something in the hash map.
type RemoveOperation struct {
	baseOperation
}

func (its *RemoveOperation) GetBody() *removeBody {
	return its.Body.(*removeBody)
}
