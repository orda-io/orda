package operations

import (
	"github.com/orda-io/orda/client/pkg/model"
)

// NewInsertOperation creates a new InsertOperation
func NewInsertOperation(pos int, values []interface{}) *InsertOperation {
	return &InsertOperation{
		baseOperation: newBaseOperation(
			model.TypeOfOperation_LIST_INSERT,
			nil,
			&InsertBody{
				V: values,
			},
		),
		Pos: pos,
	}
}

// InsertBody is the body of InsertOperation
type InsertBody struct {
	T *model.Timestamp
	V []interface{}
}

// InsertOperation is used to insert a value to a list
type InsertOperation struct {
	baseOperation
	Pos int // for local
}

// GetBody returns the body
func (its *InsertOperation) GetBody() *InsertBody {
	return its.Body.(*InsertBody)
}

// ////////////////// DeleteOperation ////////////////////

// NewDeleteOperation creates a new DeleteOperation.
func NewDeleteOperation(pos int, numOfNodes int) *DeleteOperation {
	return &DeleteOperation{
		baseOperation: newBaseOperation(
			model.TypeOfOperation_LIST_DELETE,
			nil,
			&DeleteBody{},
		),
		Pos:        pos,
		NumOfNodes: numOfNodes,
	}
}

// DeleteBody is the body of DeleteOperation
type DeleteBody struct {
	T []*model.Timestamp
}

// DeleteOperation is used to delete a value from a list.
type DeleteOperation struct {
	baseOperation
	Pos        int
	NumOfNodes int
}

// GetBody returns the body
func (its *DeleteOperation) GetBody() *DeleteBody {
	return its.Body.(*DeleteBody)
}

// ////////////////// UpdateOperation ////////////////////

// NewUpdateOperation creates a new UpdateOperation.
func NewUpdateOperation(pos int, values []interface{}) *UpdateOperation {
	return &UpdateOperation{
		baseOperation: newBaseOperation(
			model.TypeOfOperation_LIST_UPDATE,
			nil,
			&UpdateBody{
				V: values,
			},
		),
		Pos: pos,
	}
}

// UpdateBody is the body of UpdateOperation
type UpdateBody struct {
	T []*model.Timestamp
	V []interface{}
}

// UpdateOperation is used to update a value in a list.
type UpdateOperation struct {
	baseOperation
	Pos int
}

// GetBody returns the body
func (its *UpdateOperation) GetBody() *UpdateBody {
	return its.Body.(*UpdateBody)
}
