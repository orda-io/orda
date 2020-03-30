package operations

import "github.com/knowhunger/ortoo/ortoo/model"

func NewInsertOperation(pos int, values ...interface{}) *InsertOperation {
	return &InsertOperation{
		baseOperation: newBaseOperation(nil),
		Pos:           int32(pos),
		C: insertContent{
			Values: values,
		},
	}
}

type insertContent struct {
	Target *model.Timestamp
	Values []interface{}
}

type InsertOperation struct {
	*baseOperation
	Pos int32
	C   insertContent
}

func (its *InsertOperation) ExecuteLocal(datatype model.Datatype) (interface{}, error) {
	return datatype.ExecuteLocal(its)
}

func (its *InsertOperation) ExecuteRemote(datatype model.Datatype) (interface{}, error) {
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
		Pos:           int32(pos),
		NumOfNodes:    int32(numOfNodes),
		C:             deleteContent{},
	}
}

type deleteContent struct {
	Targets []*model.Timestamp
}

type DeleteOperation struct {
	*baseOperation
	Pos        int32
	NumOfNodes int32
	C          deleteContent
}

func (its *DeleteOperation) ExecuteLocal(datatype model.Datatype) (interface{}, error) {
	return datatype.ExecuteLocal(its)
}

func (its *DeleteOperation) ExecuteRemote(datatype model.Datatype) (interface{}, error) {
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

type updateContent struct {
}

type UpdateOperation struct {
	*baseOperation
	C updateContent
}

func (its *UpdateOperation) ExecuteLocal(datatype model.Datatype) (interface{}, error) {
	return datatype.ExecuteLocal(its)
}

func (its *UpdateOperation) ExecuteRemote(datatype model.Datatype) (interface{}, error) {
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
