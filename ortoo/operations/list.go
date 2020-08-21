package operations

import (
	"github.com/knowhunger/ortoo/ortoo/errors"
	"github.com/knowhunger/ortoo/ortoo/iface"
	"github.com/knowhunger/ortoo/ortoo/model"
)

func NewInsertOperation(pos int, values []interface{}) *InsertOperation {
	return &InsertOperation{
		baseOperation: newBaseOperation(nil),
		Pos:           pos,
		C: insertContent{
			V: values,
		},
	}
}

type insertContent struct {
	T *model.Timestamp
	V []interface{}
}

type InsertOperation struct {
	*baseOperation
	Pos int // for local
	C   insertContent
}

func (its *InsertOperation) ExecuteLocal(datatype iface.Datatype) (interface{}, errors.OrtooError) {
	return datatype.ExecuteLocal(its)
}

func (its *InsertOperation) ExecuteRemote(datatype iface.Datatype) (interface{}, errors.OrtooError) {
	return datatype.ExecuteRemote(its)
}

func (its *InsertOperation) ToModelOperation() *model.Operation {
	return &model.Operation{
		ID:     its.ID,
		OpType: model.TypeOfOperation_LIST_INSERT,
		Json:   marshalContent(its.C),
	}
}

func (its *InsertOperation) GetType() model.TypeOfOperation {
	return model.TypeOfOperation_LIST_INSERT
}

func (its *InsertOperation) String() string {
	return toString(its.ID, its.C)
}

// ////////////////// DeleteOperation ////////////////////

func NewDeleteOperation(pos int, numOfNodes int) *DeleteOperation {
	return &DeleteOperation{
		baseOperation: newBaseOperation(nil),
		Pos:           pos,
		NumOfNodes:    numOfNodes,
		C:             deleteContent{},
	}
}

type deleteContent struct {
	T []*model.Timestamp
}

type DeleteOperation struct {
	*baseOperation
	Pos        int
	NumOfNodes int
	C          deleteContent
}

func (its *DeleteOperation) ExecuteLocal(datatype iface.Datatype) (interface{}, errors.OrtooError) {
	return datatype.ExecuteLocal(its)
}

func (its *DeleteOperation) ExecuteRemote(datatype iface.Datatype) (interface{}, errors.OrtooError) {
	return datatype.ExecuteRemote(its)
}

func (its *DeleteOperation) ToModelOperation() *model.Operation {
	return &model.Operation{
		ID:     its.ID,
		OpType: model.TypeOfOperation_LIST_DELETE,
		Json:   marshalContent(its.C),
	}
}

func (its *DeleteOperation) GetType() model.TypeOfOperation {
	return model.TypeOfOperation_LIST_DELETE
}

func (its *DeleteOperation) String() string {
	return toString(its.ID, its.C)
}

// ////////////////// UpdateOperation ////////////////////

func NewUpdateOperation(pos int, values []interface{}) *UpdateOperation {
	return &UpdateOperation{
		baseOperation: newBaseOperation(nil),
		Pos:           pos,
		C: updateContent{
			V: values,
		},
	}
}

type updateContent struct {
	T []*model.Timestamp
	V []interface{}
}

type UpdateOperation struct {
	*baseOperation
	Pos int
	C   updateContent
}

func (its *UpdateOperation) ExecuteLocal(datatype iface.Datatype) (interface{}, errors.OrtooError) {
	return datatype.ExecuteLocal(its)
}

func (its *UpdateOperation) ExecuteRemote(datatype iface.Datatype) (interface{}, errors.OrtooError) {
	return datatype.ExecuteRemote(its)
}

func (its *UpdateOperation) ToModelOperation() *model.Operation {
	return &model.Operation{
		ID:     its.ID,
		OpType: model.TypeOfOperation_LIST_UPDATE,
		Json:   marshalContent(its.C),
	}
}

func (its *UpdateOperation) GetType() model.TypeOfOperation {
	return model.TypeOfOperation_LIST_UPDATE
}

func (its *UpdateOperation) String() string {
	return toString(its.ID, its.C)
}
