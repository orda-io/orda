package operations

import (
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/iface"
	"github.com/knowhunger/ortoo/pkg/model"
	"strings"
)

// ////////////////// DocPutInObjectOperation ////////////////////

// NewDocPutInObjectOperation creates a new DocPutInObjectOperation.
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

// DocPutInObjectOperation is used to put a value into JSONObject.
type DocPutInObjectOperation struct {
	*baseOperation
	C docPutInObjectContent
}

// ExecuteLocal enables the operation to perform something at the local client.
func (its *DocPutInObjectOperation) ExecuteLocal(datatype iface.Datatype) (interface{}, errors.OrtooError) {
	return datatype.ExecuteLocal(its)
}

// ExecuteRemote enables the operation to perform something at the remote clients.
func (its *DocPutInObjectOperation) ExecuteRemote(datatype iface.Datatype) (interface{}, errors.OrtooError) {
	return datatype.ExecuteRemote(its)
}

// ToModelOperation transforms this operation to the model.Operation.
func (its *DocPutInObjectOperation) ToModelOperation() *model.Operation {
	return &model.Operation{
		ID:     its.ID,
		OpType: model.TypeOfOperation_DOCUMENT_PUT_OBJ,
		Json:   marshalContent(its.C),
	}
}

// GetType returns the type of operation.
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

// NewDocInsertToArrayOperation creates a new DocInsertToArrayOperation.
func NewDocInsertToArrayOperation(parent *model.Timestamp, pos int, values []interface{}) *DocInsertToArrayOperation {
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

// DocInsertToArrayOperation is used to put a value into JSONArray.
type DocInsertToArrayOperation struct {
	*baseOperation
	Pos int
	C   docInsertToArrayContent
}

// ExecuteLocal enables the operation to perform something at the local client.
func (its *DocInsertToArrayOperation) ExecuteLocal(datatype iface.Datatype) (interface{}, errors.OrtooError) {
	return datatype.ExecuteLocal(its)
}

// ExecuteRemote enables the operation to perform something at the remote clients.
func (its *DocInsertToArrayOperation) ExecuteRemote(datatype iface.Datatype) (interface{}, errors.OrtooError) {
	return datatype.ExecuteRemote(its)
}

// ToModelOperation transforms this operation to the model.Operation.
func (its *DocInsertToArrayOperation) ToModelOperation() *model.Operation {
	return &model.Operation{
		ID:     its.ID,
		OpType: model.TypeOfOperation_DOCUMENT_INS_ARR,
		Json:   marshalContent(its.C),
	}
}

// GetType returns the type of operation.
func (its *DocInsertToArrayOperation) GetType() model.TypeOfOperation {
	return model.TypeOfOperation_DOCUMENT_INS_ARR
}

// GetType returns the type of operation.
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

// NewDocDeleteInObjectOperation creates a new DocDeleteInObjectOperation.
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

// DocDeleteInObjectOperation is used to delete a value from JSONObject.
type DocDeleteInObjectOperation struct {
	*baseOperation
	C docDeleteInObjectContent
}

// ExecuteLocal enables the operation to perform something at the local client.
func (its *DocDeleteInObjectOperation) ExecuteLocal(datatype iface.Datatype) (interface{}, errors.OrtooError) {
	return datatype.ExecuteLocal(its)
}

// ExecuteRemote enables the operation to perform something at the remote clients.
func (its *DocDeleteInObjectOperation) ExecuteRemote(datatype iface.Datatype) (interface{}, errors.OrtooError) {
	return datatype.ExecuteRemote(its)
}

// ToModelOperation transforms this operation to the model.Operation.
func (its *DocDeleteInObjectOperation) ToModelOperation() *model.Operation {
	return &model.Operation{
		ID:     its.ID,
		OpType: model.TypeOfOperation_DOCUMENT_DEL_OBJ,
		Json:   marshalContent(its.C),
	}
}

// GetType returns the type of operation.
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

// NewDocUpdateInArrayOperation creates a new DocUpdateInArrayOperation.
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

// DocUpdateInArrayOperation is used to update a value into JSONArray.
type DocUpdateInArrayOperation struct {
	*baseOperation
	Pos int // for local
	C   docUpdateInArrayContent
}

// ExecuteLocal enables the operation to perform something at the local client.
func (its *DocUpdateInArrayOperation) ExecuteLocal(datatype iface.Datatype) (interface{}, errors.OrtooError) {
	return datatype.ExecuteLocal(its)
}

// ExecuteRemote enables the operation to perform something at the remote clients.
func (its *DocUpdateInArrayOperation) ExecuteRemote(datatype iface.Datatype) (interface{}, errors.OrtooError) {
	return datatype.ExecuteRemote(its)
}

// ToModelOperation transforms this operation to the model.Operation.
func (its *DocUpdateInArrayOperation) ToModelOperation() *model.Operation {
	return &model.Operation{
		ID:     its.ID,
		OpType: model.TypeOfOperation_DOCUMENT_UPD_ARR,
		Json:   marshalContent(its.C),
	}
}

// GetType returns the type of operation.
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

// NewDocDeleteInArrayOperation creates a new DocDeleteInArrayOperation.
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

// DocDeleteInArrayOperation is used to delete a value into JSONArray.
type DocDeleteInArrayOperation struct {
	*baseOperation
	Pos        int // for local
	NumOfNodes int // for local
	C          docDeleteInArrayContent
}

// ExecuteLocal enables the operation to perform something at the local client.
func (its *DocDeleteInArrayOperation) ExecuteLocal(datatype iface.Datatype) (interface{}, errors.OrtooError) {
	return datatype.ExecuteLocal(its)
}

// ExecuteRemote enables the operation to perform something at the remote clients.
func (its *DocDeleteInArrayOperation) ExecuteRemote(datatype iface.Datatype) (interface{}, errors.OrtooError) {
	return datatype.ExecuteRemote(its)
}

// ToModelOperation transforms this operation to the model.Operation.
func (its *DocDeleteInArrayOperation) ToModelOperation() *model.Operation {
	return &model.Operation{
		ID:     its.ID,
		OpType: model.TypeOfOperation_DOCUMENT_DEL_ARR,
		Json:   marshalContent(its.C),
	}
}

// GetType returns the type of operation.
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
