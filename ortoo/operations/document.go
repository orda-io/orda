package operations

import (
	"github.com/knowhunger/ortoo/ortoo/errors"
	"github.com/knowhunger/ortoo/ortoo/iface"
	"github.com/knowhunger/ortoo/ortoo/model"
	"strings"
)

// ////////////////// DocPutInObjectOperation ////////////////////

func NewDocPutInObjectOperation(parent *model.Timestamp, key string, value interface{}) *DocPutInObjectOperation {
	return &DocPutInObjectOperation{
		baseOperation: newBaseOperation(nil),
		C: docPutInObjectContent{
			P: parent,
			K: key,
			V: value,
		},
	}
}

type docPutInObjectContent struct {
	P *model.Timestamp
	K string
	V interface{}
}

type DocPutInObjectOperation struct {
	*baseOperation
	C docPutInObjectContent
}

func (its *DocPutInObjectOperation) ExecuteLocal(datatype iface.Datatype) (interface{}, errors.OrtooError) {
	return datatype.ExecuteLocal(its)
}

func (its *DocPutInObjectOperation) ExecuteRemote(datatype iface.Datatype) (interface{}, errors.OrtooError) {
	return datatype.ExecuteRemote(its)
}

func (its *DocPutInObjectOperation) ToModelOperation() *model.Operation {
	return &model.Operation{
		ID:     its.ID,
		OpType: model.TypeOfOperation_DOCUMENT_PUT_OBJ,
		Json:   marshalContent(its.C),
	}
}

func (its *DocPutInObjectOperation) GetType() model.TypeOfOperation {
	return model.TypeOfOperation_DOCUMENT_PUT_OBJ
}

func (its *DocPutInObjectOperation) String() string {
	var sb strings.Builder
	sb.WriteString(its.GetType().String())
	sb.WriteString("[ID")
	sb.WriteString(its.ID.ToString())
	sb.WriteString(":")
	sb.WriteString("]")
	return sb.String()
}

// ////////////////// DocInsertToArrayOperation ////////////////////

func NewDocInsToArrayOperation(parent *model.Timestamp, pos int, values []interface{}) *DocInsertToArrayOperation {
	return &DocInsertToArrayOperation{
		baseOperation: newBaseOperation(nil),
		Pos:           pos,
		C: docInsertToArrayContent{
			P: parent,
			V: values,
		},
	}
}

type docInsertToArrayContent struct {
	P *model.Timestamp
	T *model.Timestamp
	V []interface{}
}

type DocInsertToArrayOperation struct {
	*baseOperation
	Pos int
	C   docInsertToArrayContent
}

func (its *DocInsertToArrayOperation) ExecuteLocal(datatype iface.Datatype) (interface{}, errors.OrtooError) {
	return datatype.ExecuteLocal(its)
}

func (its *DocInsertToArrayOperation) ExecuteRemote(datatype iface.Datatype) (interface{}, errors.OrtooError) {
	return datatype.ExecuteRemote(its)
}

func (its *DocInsertToArrayOperation) ToModelOperation() *model.Operation {
	return &model.Operation{
		ID:     its.ID,
		OpType: model.TypeOfOperation_DOCUMENT_INS_ARR,
		Json:   marshalContent(its.C),
	}
}

func (its *DocInsertToArrayOperation) GetType() model.TypeOfOperation {
	return model.TypeOfOperation_DOCUMENT_INS_ARR
}

func (its *DocInsertToArrayOperation) String() string {
	var sb strings.Builder
	sb.WriteString(its.GetType().String())
	sb.WriteString("[")
	sb.WriteString(its.C.T.ToString())
	sb.WriteString(":")

	sb.WriteString("]")
	return sb.String()
}

// ////////////////// DocDeleteInObjectOperation ////////////////////

func NewDocDeleteInObjectOperation(parent *model.Timestamp, key string) *DocDeleteInObjectOperation {
	return &DocDeleteInObjectOperation{
		baseOperation: newBaseOperation(nil),
		C: docDeleteInObjectContent{
			P:   parent,
			Key: key,
		},
	}
}

type docDeleteInObjectContent struct {
	P   *model.Timestamp
	Key string
}

type DocDeleteInObjectOperation struct {
	*baseOperation
	C docDeleteInObjectContent
}

func (its *DocDeleteInObjectOperation) ExecuteLocal(datatype iface.Datatype) (interface{}, errors.OrtooError) {
	return datatype.ExecuteLocal(its)
}

func (its *DocDeleteInObjectOperation) ExecuteRemote(datatype iface.Datatype) (interface{}, errors.OrtooError) {
	return datatype.ExecuteRemote(its)
}

func (its *DocDeleteInObjectOperation) ToModelOperation() *model.Operation {
	return &model.Operation{
		ID:     its.ID,
		OpType: model.TypeOfOperation_DOCUMENT_DEL_OBJ,
		Json:   marshalContent(its.C),
	}
}

func (its *DocDeleteInObjectOperation) GetType() model.TypeOfOperation {
	return model.TypeOfOperation_DOCUMENT_DEL_OBJ
}

func (its *DocDeleteInObjectOperation) String() string {
	var sb strings.Builder
	sb.WriteString(its.GetType().String())
	sb.WriteString("[")
	sb.WriteString("]")
	return sb.String()
}

// ////////////////// UpdInObjectOperation ////////////////////

func NewDocUpdateInArrayOperation(parent *model.Timestamp, pos int, values []interface{}) *DocUpdateInArrayOperation {
	return &DocUpdateInArrayOperation{
		baseOperation: newBaseOperation(nil),
		Pos:           pos,
		C: docUpdateInArrayContent{
			P: parent,
			V: values,
		},
	}
}

type docUpdateInArrayContent struct {
	P *model.Timestamp
	T []*model.Timestamp
	V []interface{}
}

type DocUpdateInArrayOperation struct {
	*baseOperation
	Pos int // for local
	C   docUpdateInArrayContent
}

func (its *DocUpdateInArrayOperation) ExecuteLocal(datatype iface.Datatype) (interface{}, errors.OrtooError) {
	return datatype.ExecuteLocal(its)
}

func (its *DocUpdateInArrayOperation) ExecuteRemote(datatype iface.Datatype) (interface{}, errors.OrtooError) {
	return datatype.ExecuteRemote(its)
}

func (its *DocUpdateInArrayOperation) ToModelOperation() *model.Operation {
	return &model.Operation{
		ID:     its.ID,
		OpType: model.TypeOfOperation_DOCUMENT_UPD_ARR,
		Json:   marshalContent(its.C),
	}
}

func (its *DocUpdateInArrayOperation) GetType() model.TypeOfOperation {
	return model.TypeOfOperation_DOCUMENT_UPD_ARR
}

func (its *DocUpdateInArrayOperation) String() string {
	var sb strings.Builder
	sb.WriteString(its.GetType().String())
	sb.WriteString("[")

	sb.WriteString("]")
	return sb.String()
}

// ////////////////// DocDeleteInArrayOperation ////////////////////

func NewDocDeleteInArrayOperation(parent *model.Timestamp, pos, numOfNodes int) *DocDeleteInArrayOperation {
	return &DocDeleteInArrayOperation{
		baseOperation: newBaseOperation(nil),
		Pos:           pos,
		NumOfNodes:    numOfNodes,
		C: docDeleteInArrayContent{
			P: parent,
		},
	}
}

type docDeleteInArrayContent struct {
	P *model.Timestamp
	T []*model.Timestamp
}

type DocDeleteInArrayOperation struct {
	*baseOperation
	Pos        int // for local
	NumOfNodes int // for local
	C          docDeleteInArrayContent
}

func (its *DocDeleteInArrayOperation) ExecuteLocal(datatype iface.Datatype) (interface{}, errors.OrtooError) {
	return datatype.ExecuteLocal(its)
}

func (its *DocDeleteInArrayOperation) ExecuteRemote(datatype iface.Datatype) (interface{}, errors.OrtooError) {
	return datatype.ExecuteRemote(its)
}

func (its *DocDeleteInArrayOperation) ToModelOperation() *model.Operation {
	return &model.Operation{
		ID:     its.ID,
		OpType: model.TypeOfOperation_DOCUMENT_DEL_ARR,
		Json:   marshalContent(its.C),
	}
}

func (its *DocDeleteInArrayOperation) GetType() model.TypeOfOperation {
	return model.TypeOfOperation_DOCUMENT_DEL_ARR
}

func (its *DocDeleteInArrayOperation) String() string {
	var sb strings.Builder
	sb.WriteString(its.GetType().String())
	sb.WriteString("[")

	sb.WriteString("]")
	return sb.String()
}
