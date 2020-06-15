package operations

import (
	"github.com/knowhunger/ortoo/ortoo/iface"
	"github.com/knowhunger/ortoo/ortoo/model"
)

// ////////////////// AddObjectOperation ////////////////////

func NewAddObjectOperation(parent *model.Timestamp, key string, value interface{}) *AddObjectOperation {
	return &AddObjectOperation{
		baseOperation: newBaseOperation(nil),
		C: addObjectContent{
			P: parent,
			K: key,
			V: value,
		},
	}
}

type addObjectContent struct {
	P *model.Timestamp
	K string
	V interface{}
}

type AddObjectOperation struct {
	*baseOperation
	C addObjectContent
}

func (its *AddObjectOperation) ExecuteLocal(datatype iface.Datatype) (interface{}, error) {
	return datatype.ExecuteLocal(its)
}

func (its *AddObjectOperation) ExecuteRemote(datatype iface.Datatype) (interface{}, error) {
	return datatype.ExecuteRemote(its)
}

func (its *AddObjectOperation) ToModelOperation() *model.Operation {
	return &model.Operation{
		ID:     its.ID,
		OpType: model.TypeOfOperation_DOCUMENT_ADD_OBJ,
		Json:   marshalContent(its.C),
	}
}

func (its *AddObjectOperation) GetType() model.TypeOfOperation {
	return model.TypeOfOperation_DOCUMENT_ADD_OBJ
}

func (its *AddObjectOperation) String() string {
	panic("implement me")
}

// ////////////////// AddArrayOperation ////////////////////

func NewAddArrayOperation(parent *model.Timestamp, pos int, values []interface{}) *AddArrayOperation {
	return &AddArrayOperation{
		baseOperation: newBaseOperation(nil),
		Pos:           int32(pos),
		C: addArrayContent{
			P: parent,
			V: values,
		},
	}
}

type addArrayContent struct {
	P *model.Timestamp
	T *model.Timestamp
	V []interface{}
}

type AddArrayOperation struct {
	*baseOperation
	Pos int32
	C   addArrayContent
}

func (its *AddArrayOperation) ExecuteLocal(datatype iface.Datatype) (interface{}, error) {
	return datatype.ExecuteLocal(its)
}

func (its *AddArrayOperation) ExecuteRemote(datatype iface.Datatype) (interface{}, error) {
	return datatype.ExecuteRemote(its)
}

func (its *AddArrayOperation) ToModelOperation() *model.Operation {
	return &model.Operation{
		ID:     its.ID,
		OpType: model.TypeOfOperation_DOCUMENT_ADD_ARR,
		Json:   marshalContent(its.C),
	}
}

func (its *AddArrayOperation) GetType() model.TypeOfOperation {
	return model.TypeOfOperation_DOCUMENT_ADD_ARR
}

func (its *AddArrayOperation) String() string {
	panic("implement me")
}

// ////////////////// DeleteInObjectOperation ////////////////////

func NewDeleteInObjectOperation(parent *model.Timestamp, key string) *DeleteInObjectOperation {
	return &DeleteInObjectOperation{
		baseOperation: newBaseOperation(nil),
		C: deleteInObjectContent{
			P:   parent,
			Key: key,
		},
	}
}

type deleteInObjectContent struct {
	P   *model.Timestamp
	Key string
}

type DeleteInObjectOperation struct {
	*baseOperation
	C deleteInObjectContent
}

func (its *DeleteInObjectOperation) ExecuteLocal(datatype iface.Datatype) (interface{}, error) {
	return datatype.ExecuteLocal(its)
}

func (its *DeleteInObjectOperation) ExecuteRemote(datatype iface.Datatype) (interface{}, error) {
	return datatype.ExecuteRemote(its)
}

func (its *DeleteInObjectOperation) ToModelOperation() *model.Operation {
	return &model.Operation{
		ID:     its.ID,
		OpType: model.TypeOfOperation_DOCUMENT_DEL_OBJ,
		Json:   marshalContent(its.C),
	}
}

func (its *DeleteInObjectOperation) GetType() model.TypeOfOperation {
	return model.TypeOfOperation_DOCUMENT_DEL_OBJ
}

func (its *DeleteInObjectOperation) String() string {
	panic("implement me")
}

// ////////////////// DeleteInObjectOperation ////////////////////

func NewDeleteInArrayOperation(parent *model.Timestamp, pos, numOfNodes int) *DeleteInArrayOperation {
	return &DeleteInArrayOperation{
		baseOperation: newBaseOperation(nil),
		Pos:           int32(pos),
		NumOfNodes:    int32(numOfNodes),
		C: deleteInArrayContent{
			P: parent,
		},
	}
}

type deleteInArrayContent struct {
	P *model.Timestamp
	T []*model.Timestamp
}

type DeleteInArrayOperation struct {
	*baseOperation
	Pos        int
	NumOfNodes int
	C          deleteInArrayContent
}

func (its *DeleteInArrayOperation) ExecuteLocal(datatype iface.Datatype) (interface{}, error) {
	return datatype.ExecuteLocal(its)
}

func (its *DeleteInArrayOperation) ExecuteRemote(datatype iface.Datatype) (interface{}, error) {
	return datatype.ExecuteRemote(its)
}

func (its *DeleteInArrayOperation) ToModelOperation() *model.Operation {
	return &model.Operation{
		ID:     its.ID,
		OpType: model.TypeOfOperation_DOCUMENT_DEL_OBJ,
		Json:   marshalContent(its.C),
	}
}

func (its *DeleteInArrayOperation) GetType() model.TypeOfOperation {
	return model.TypeOfOperation_DOCUMENT_DEL_OBJ
}

func (its *DeleteInArrayOperation) String() string {
	panic("implement me")
}

// ////////////////// SetOperation ////////////////////

type SetOperation struct {
	*baseOperation
}

func (its *SetOperation) ExecuteLocal(datatype iface.Datatype) (interface{}, error) {
	return datatype.ExecuteLocal(its)
}

func (its *SetOperation) ExecuteRemote(datatype iface.Datatype) (interface{}, error) {
	return datatype.ExecuteRemote(its)
}

func (its *SetOperation) ToModelOperation() *model.Operation {
	panic("implement me")
}

func (its *SetOperation) GetType() model.TypeOfOperation {
	return model.TypeOfOperation_DOCUMENT_SET
}

func (its *SetOperation) String() string {
	panic("implement me")
}
