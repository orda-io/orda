package operations

import (
	model2 "github.com/orda-io/orda/client/pkg/model"
)

// NewInsertOperation creates a new InsertOperation
func NewInsertOperation(pos int, values []interface{}) *InsertOperation {
	return &InsertOperation{
		baseOperation: newBaseOperation(
			model2.TypeOfOperation_LIST_INSERT,
			nil,
			&insertBody{
				V: values,
			},
		),
		Pos: pos,
	}
}

type insertBody struct {
	T *model2.Timestamp
	V []interface{}
}

// InsertOperation is used to insert a value to a list
type InsertOperation struct {
	baseOperation
	Pos int // for local
}

func (its *InsertOperation) GetBody() *insertBody {
	return its.Body.(*insertBody)
}

// ////////////////// DeleteOperation ////////////////////

// NewDeleteOperation creates a new DeleteOperation.
func NewDeleteOperation(pos int, numOfNodes int) *DeleteOperation {
	return &DeleteOperation{
		baseOperation: newBaseOperation(
			model2.TypeOfOperation_LIST_DELETE,
			nil,
			&deleteBody{},
		),
		Pos:        pos,
		NumOfNodes: numOfNodes,
	}
}

type deleteBody struct {
	T []*model2.Timestamp
}

// DeleteOperation is used to delete a value from a list.
type DeleteOperation struct {
	baseOperation
	Pos        int
	NumOfNodes int
}

func (its *DeleteOperation) GetBody() *deleteBody {
	return its.Body.(*deleteBody)
}

// ////////////////// UpdateOperation ////////////////////

// NewUpdateOperation creates a new UpdateOperation.
func NewUpdateOperation(pos int, values []interface{}) *UpdateOperation {
	return &UpdateOperation{
		baseOperation: newBaseOperation(
			model2.TypeOfOperation_LIST_UPDATE,
			nil,
			&updateBody{
				V: values,
			},
		),
		Pos: pos,
	}
}

type updateBody struct {
	T []*model2.Timestamp
	V []interface{}
}

// UpdateOperation is used to update a value in a list.
type UpdateOperation struct {
	baseOperation
	Pos int
}

func (its *UpdateOperation) GetBody() *updateBody {
	return its.Body.(*updateBody)
}
