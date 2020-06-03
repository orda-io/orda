package operations

import (
	"github.com/knowhunger/ortoo/ortoo/iface"
	"github.com/knowhunger/ortoo/ortoo/model"
)

// ////////////////// AddOperation ////////////////////

func NewAddOperation(parent *model.Timestamp, key string, value interface{}) *AddOperation {
	return &AddOperation{
		baseOperation: newBaseOperation(nil),
		C: addContent{
			P: parent,
			K: key,
			V: value,
		},
	}
}

type addContent struct {
	P *model.Timestamp
	K string
	V interface{}
}

type AddOperation struct {
	*baseOperation
	C addContent
}

func (its *AddOperation) ExecuteLocal(datatype iface.Datatype) (interface{}, error) {
	return datatype.ExecuteLocal(its)
}

func (its *AddOperation) ExecuteRemote(datatype iface.Datatype) (interface{}, error) {
	return datatype.ExecuteRemote(its)
}

func (its *AddOperation) ToModelOperation() *model.Operation {
	return &model.Operation{
		ID:     its.ID,
		OpType: model.TypeOfOperation_DOCUMENT_ADD,
		Json:   marshalContent(its.C),
	}
}

func (its *AddOperation) GetType() model.TypeOfOperation {
	return model.TypeOfOperation_DOCUMENT_ADD
}

func (its *AddOperation) String() string {
	panic("implement me")
}

// ////////////////// CutOperation ////////////////////

type CutOperation struct {
	*baseOperation
}

func (its *CutOperation) ExecuteLocal(datatype iface.Datatype) (interface{}, error) {
	return datatype.ExecuteLocal(its)
}

func (its *CutOperation) ExecuteRemote(datatype iface.Datatype) (interface{}, error) {
	return datatype.ExecuteRemote(its)
}

func (its *CutOperation) ToModelOperation() *model.Operation {
	panic("implement me")
}

func (its *CutOperation) GetType() model.TypeOfOperation {
	return model.TypeOfOperation_DOCUMENT_CUT
}

func (its *CutOperation) String() string {
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
