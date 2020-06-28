package operations

import (
	"github.com/knowhunger/ortoo/ortoo/iface"
	"github.com/knowhunger/ortoo/ortoo/model"
	"strings"
)

// ////////////////// PutObjectOperation ////////////////////

func NewPutObjectOperation(parent *model.Timestamp, key string, value interface{}) *PutObjectOperation {
	return &PutObjectOperation{
		baseOperation: newBaseOperation(nil),
		C: putObjectContent{
			P: parent,
			K: key,
			V: value,
		},
	}
}

type putObjectContent struct {
	P *model.Timestamp
	K string
	V interface{}
}

type PutObjectOperation struct {
	*baseOperation
	C putObjectContent
}

func (its *PutObjectOperation) ExecuteLocal(datatype iface.Datatype) (interface{}, error) {
	return datatype.ExecuteLocal(its)
}

func (its *PutObjectOperation) ExecuteRemote(datatype iface.Datatype) (interface{}, error) {
	return datatype.ExecuteRemote(its)
}

func (its *PutObjectOperation) ToModelOperation() *model.Operation {
	return &model.Operation{
		ID:     its.ID,
		OpType: model.TypeOfOperation_DOCUMENT_PUT_OBJ,
		Json:   marshalContent(its.C),
	}
}

func (its *PutObjectOperation) GetType() model.TypeOfOperation {
	return model.TypeOfOperation_DOCUMENT_PUT_OBJ
}

func (its *PutObjectOperation) String() string {
	var sb strings.Builder
	sb.WriteString(its.GetType().String())
	sb.WriteString("[ID")
	sb.WriteString(its.ID.ToString())
	sb.WriteString(":")
	sb.WriteString("]")
	return sb.String()
}

// ////////////////// InsArrayOperation ////////////////////

func NewInsArrayOperation(parent *model.Timestamp, pos int, values []interface{}) *InsArrayOperation {
	return &InsArrayOperation{
		baseOperation: newBaseOperation(nil),
		Pos:           pos,
		C: insArrayContent{
			P: parent,
			V: values,
		},
	}
}

type insArrayContent struct {
	P *model.Timestamp
	T *model.Timestamp
	V []interface{}
}

type InsArrayOperation struct {
	*baseOperation
	Pos int
	C   insArrayContent
}

func (its *InsArrayOperation) ExecuteLocal(datatype iface.Datatype) (interface{}, error) {
	return datatype.ExecuteLocal(its)
}

func (its *InsArrayOperation) ExecuteRemote(datatype iface.Datatype) (interface{}, error) {
	return datatype.ExecuteRemote(its)
}

func (its *InsArrayOperation) ToModelOperation() *model.Operation {
	return &model.Operation{
		ID:     its.ID,
		OpType: model.TypeOfOperation_DOCUMENT_INS_ARR,
		Json:   marshalContent(its.C),
	}
}

func (its *InsArrayOperation) GetType() model.TypeOfOperation {
	return model.TypeOfOperation_DOCUMENT_INS_ARR
}

func (its *InsArrayOperation) String() string {
	var sb strings.Builder
	sb.WriteString(its.GetType().String())
	sb.WriteString("[")
	sb.WriteString(its.C.T.ToString())
	sb.WriteString(":")

	sb.WriteString("]")
	return sb.String()
}

// ////////////////// DelInObjectOperation ////////////////////

func NewDelInObjectOperation(parent *model.Timestamp, key string) *DelInObjectOperation {
	return &DelInObjectOperation{
		baseOperation: newBaseOperation(nil),
		C: delInObjectContent{
			P:   parent,
			Key: key,
		},
	}
}

type delInObjectContent struct {
	P   *model.Timestamp
	Key string
}

type DelInObjectOperation struct {
	*baseOperation
	C delInObjectContent
}

func (its *DelInObjectOperation) ExecuteLocal(datatype iface.Datatype) (interface{}, error) {
	return datatype.ExecuteLocal(its)
}

func (its *DelInObjectOperation) ExecuteRemote(datatype iface.Datatype) (interface{}, error) {
	return datatype.ExecuteRemote(its)
}

func (its *DelInObjectOperation) ToModelOperation() *model.Operation {
	return &model.Operation{
		ID:     its.ID,
		OpType: model.TypeOfOperation_DOCUMENT_DEL_OBJ,
		Json:   marshalContent(its.C),
	}
}

func (its *DelInObjectOperation) GetType() model.TypeOfOperation {
	return model.TypeOfOperation_DOCUMENT_DEL_OBJ
}

func (its *DelInObjectOperation) String() string {
	var sb strings.Builder
	sb.WriteString(its.GetType().String())
	sb.WriteString("[")
	sb.WriteString("]")
	return sb.String()
}

// ////////////////// DelInObjectOperation ////////////////////

func NewDelInArrayOperation(parent *model.Timestamp, pos, numOfNodes int) *DelInArrayOperation {
	return &DelInArrayOperation{
		baseOperation: newBaseOperation(nil),
		Pos:           pos,
		NumOfNodes:    numOfNodes,
		C: delInArrayContent{
			P: parent,
		},
	}
}

type delInArrayContent struct {
	P *model.Timestamp
	T []*model.Timestamp
}

type DelInArrayOperation struct {
	*baseOperation
	Pos        int
	NumOfNodes int
	C          delInArrayContent
}

func (its *DelInArrayOperation) ExecuteLocal(datatype iface.Datatype) (interface{}, error) {
	return datatype.ExecuteLocal(its)
}

func (its *DelInArrayOperation) ExecuteRemote(datatype iface.Datatype) (interface{}, error) {
	return datatype.ExecuteRemote(its)
}

func (its *DelInArrayOperation) ToModelOperation() *model.Operation {
	return &model.Operation{
		ID:     its.ID,
		OpType: model.TypeOfOperation_DOCUMENT_DEL_ARR,
		Json:   marshalContent(its.C),
	}
}

func (its *DelInArrayOperation) GetType() model.TypeOfOperation {
	return model.TypeOfOperation_DOCUMENT_DEL_ARR
}

func (its *DelInArrayOperation) String() string {
	var sb strings.Builder
	sb.WriteString(its.GetType().String())
	sb.WriteString("[")

	sb.WriteString("]")
	return sb.String()
}
