package operations

import "github.com/knowhunger/ortoo/ortoo/model"

func NewInsertOperation(pos int32, values ...interface{}) *InsertOperation {
	return &InsertOperation{
		baseOperation: newBaseOperation(nil),
		Pos:           pos,
		C: insertContent{
			Values: values,
		},
	}
}

type insertContent struct {
	Target *model.Timestamp
	Values interface{}
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
	panic("implement me")
}

func (its *InsertOperation) ToModelOperation() *model.Operation {
	panic("implement me")
}

func (its *InsertOperation) GetType() model.TypeOfOperation {
	panic("implement me")
}

func (its *InsertOperation) String() string {
	panic("implement me")
}
