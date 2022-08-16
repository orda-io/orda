package operations

import (
	"github.com/orda-io/orda/client/pkg/model"
)

// ////////////////// DocPutInObjOperation ////////////////////

// NewDocPutInObjOperation creates a new DocPutInObjOperation.
func NewDocPutInObjOperation(parent *model.Timestamp, key string, value interface{}) *DocPutInObjOperation {
	return &DocPutInObjOperation{
		baseOperation: newBaseOperation(
			model.TypeOfOperation_DOC_OBJ_PUT,
			nil,
			&DocPutInObjBody{
				P: parent,
				K: key,
				V: value,
			},
		),
	}
}

// DocPutInObjBody is the body of DocPutInObjOperation
type DocPutInObjBody struct {
	P *model.Timestamp
	K string
	V interface{}
}

// DocPutInObjOperation is used to put a value into JSONObject.
type DocPutInObjOperation struct {
	baseOperation
}

// GetBody returns the body
func (its *DocPutInObjOperation) GetBody() *DocPutInObjBody {
	return its.Body.(*DocPutInObjBody)
}

// ////////////////// DocRemoveInObjOperation ////////////////////

// NewDocRemoveInObjOperation creates a new DocRemoveInObjOperation.
func NewDocRemoveInObjOperation(parent *model.Timestamp, key string) *DocRemoveInObjOperation {
	return &DocRemoveInObjOperation{
		baseOperation: newBaseOperation(model.TypeOfOperation_DOC_OBJ_RMV, nil, &DocRemoveInObjectBody{
			P: parent,
			K: key,
		}),
	}
}

// DocRemoveInObjectBody is the body of DocRemoveInObjOperation
type DocRemoveInObjectBody struct {
	P *model.Timestamp
	K string
}

// DocRemoveInObjOperation is used to delete a value from JSONObject.
type DocRemoveInObjOperation struct {
	baseOperation
}

// GetBody returns the body
func (its *DocRemoveInObjOperation) GetBody() *DocRemoveInObjectBody {
	return its.Body.(*DocRemoveInObjectBody)
}

// ////////////////// DocInsertToArrayOperation ////////////////////

// NewDocInsertToArrayOperation creates a new DocInsertToArrayOperation.
func NewDocInsertToArrayOperation(parent *model.Timestamp, pos int, values []interface{}) *DocInsertToArrayOperation {
	return &DocInsertToArrayOperation{
		baseOperation: newBaseOperation(
			model.TypeOfOperation_DOC_ARR_INS,
			nil,
			&DocInsertToArrayBody{
				P: parent,
				V: values,
			},
		),
		Pos: pos,
	}
}

// DocInsertToArrayBody is the body of DocInsertToArrayOperation
type DocInsertToArrayBody struct {
	P *model.Timestamp
	T *model.Timestamp
	V []interface{}
}

// DocInsertToArrayOperation is used to put a value into JSONArray.
type DocInsertToArrayOperation struct {
	baseOperation
	Pos int
}

// GetBody returns the body
func (its *DocInsertToArrayOperation) GetBody() *DocInsertToArrayBody {
	return its.Body.(*DocInsertToArrayBody)
}

// ////////////////// UpdInObjectOperation ////////////////////

// NewDocUpdateInArrayOperation creates a new DocUpdateInArrayOperation.
func NewDocUpdateInArrayOperation(parent *model.Timestamp, pos int, values []interface{}) *DocUpdateInArrayOperation {
	return &DocUpdateInArrayOperation{
		baseOperation: newBaseOperation(
			model.TypeOfOperation_DOC_ARR_UPD,
			nil,
			&DocUpdateInArrayBody{
				P: parent,
				V: values,
			},
		),
		Pos: pos,
	}
}

// DocUpdateInArrayBody is the body of DocUpdateInArrayOperation
type DocUpdateInArrayBody struct {
	P *model.Timestamp
	T []*model.Timestamp
	V []interface{}
}

// DocUpdateInArrayOperation is used to update a value into JSONArray.
type DocUpdateInArrayOperation struct {
	baseOperation
	Pos int // for local
}

// GetBody returns the body
func (its *DocUpdateInArrayOperation) GetBody() *DocUpdateInArrayBody {
	return its.Body.(*DocUpdateInArrayBody)
}

// ////////////////// DocDeleteInArrayOperation ////////////////////

// NewDocDeleteInArrayOperation creates a new DocDeleteInArrayOperation.
func NewDocDeleteInArrayOperation(parent *model.Timestamp, pos, numOfNodes int) *DocDeleteInArrayOperation {
	return &DocDeleteInArrayOperation{
		baseOperation: newBaseOperation(
			model.TypeOfOperation_DOC_ARR_DEL,
			nil,
			&DocDeleteInArrayBody{
				P: parent,
			},
		),
		Pos:        pos,
		NumOfNodes: numOfNodes,
	}
}

// DocDeleteInArrayBody is the body of DocDeleteInArrayOperation
type DocDeleteInArrayBody struct {
	P *model.Timestamp
	T []*model.Timestamp
}

// DocDeleteInArrayOperation is used to delete a value into JSONArray.
type DocDeleteInArrayOperation struct {
	baseOperation
	Pos        int // for local
	NumOfNodes int // for local
}

// GetBody returns the body
func (its *DocDeleteInArrayOperation) GetBody() *DocDeleteInArrayBody {
	return its.Body.(*DocDeleteInArrayBody)
}
