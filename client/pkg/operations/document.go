package operations

import (
	model2 "github.com/orda-io/orda/client/pkg/model"
)

// ////////////////// DocPutInObjOperation ////////////////////

// NewDocPutInObjOperation creates a new DocPutInObjOperation.
func NewDocPutInObjOperation(parent *model2.Timestamp, key string, value interface{}) *DocPutInObjOperation {
	return &DocPutInObjOperation{
		baseOperation: newBaseOperation(
			model2.TypeOfOperation_DOC_OBJ_PUT,
			nil,
			&docPutInObjBody{
				P: parent,
				K: key,
				V: value,
			},
		),
	}
}

type docPutInObjBody struct {
	P *model2.Timestamp
	K string
	V interface{}
}

// DocPutInObjOperation is used to put a value into JSONObject.
type DocPutInObjOperation struct {
	baseOperation
}

func (its *DocPutInObjOperation) GetBody() *docPutInObjBody {
	return its.Body.(*docPutInObjBody)
}

// ////////////////// DocRemoveInObjOperation ////////////////////

// NewDocRemoveInObjOperation creates a new DocRemoveInObjOperation.
func NewDocRemoveInObjOperation(parent *model2.Timestamp, key string) *DocRemoveInObjOperation {
	return &DocRemoveInObjOperation{
		baseOperation: newBaseOperation(model2.TypeOfOperation_DOC_OBJ_RMV, nil, &DocRemoveInObjectBody{
			P: parent,
			K: key,
		}),
	}
}

type DocRemoveInObjectBody struct {
	P *model2.Timestamp
	K string
}

// DocRemoveInObjOperation is used to delete a value from JSONObject.
type DocRemoveInObjOperation struct {
	baseOperation
}

func (its *DocRemoveInObjOperation) GetBody() *DocRemoveInObjectBody {
	return its.Body.(*DocRemoveInObjectBody)
}

// ////////////////// DocInsertToArrayOperation ////////////////////

// NewDocInsertToArrayOperation creates a new DocInsertToArrayOperation.
func NewDocInsertToArrayOperation(parent *model2.Timestamp, pos int, values []interface{}) *DocInsertToArrayOperation {
	return &DocInsertToArrayOperation{
		baseOperation: newBaseOperation(
			model2.TypeOfOperation_DOC_ARR_INS,
			nil,
			&DocInsertToArrayBody{
				P: parent,
				V: values,
			},
		),
		Pos: pos,
	}
}

type DocInsertToArrayBody struct {
	P *model2.Timestamp
	T *model2.Timestamp
	V []interface{}
}

// DocInsertToArrayOperation is used to put a value into JSONArray.
type DocInsertToArrayOperation struct {
	baseOperation
	Pos int
}

func (its *DocInsertToArrayOperation) GetBody() *DocInsertToArrayBody {
	return its.Body.(*DocInsertToArrayBody)
}

// ////////////////// UpdInObjectOperation ////////////////////

// NewDocUpdateInArrayOperation creates a new DocUpdateInArrayOperation.
func NewDocUpdateInArrayOperation(parent *model2.Timestamp, pos int, values []interface{}) *DocUpdateInArrayOperation {
	return &DocUpdateInArrayOperation{
		baseOperation: newBaseOperation(
			model2.TypeOfOperation_DOC_ARR_UPD,
			nil,
			&DocUpdateInArrayBody{
				P: parent,
				V: values,
			},
		),
		Pos: pos,
	}
}

type DocUpdateInArrayBody struct {
	P *model2.Timestamp
	T []*model2.Timestamp
	V []interface{}
}

// DocUpdateInArrayOperation is used to update a value into JSONArray.
type DocUpdateInArrayOperation struct {
	baseOperation
	Pos int // for local
}

func (its *DocUpdateInArrayOperation) GetBody() *DocUpdateInArrayBody {
	return its.Body.(*DocUpdateInArrayBody)
}

// ////////////////// DocDeleteInArrayOperation ////////////////////

// NewDocDeleteInArrayOperation creates a new DocDeleteInArrayOperation.
func NewDocDeleteInArrayOperation(parent *model2.Timestamp, pos, numOfNodes int) *DocDeleteInArrayOperation {
	return &DocDeleteInArrayOperation{
		baseOperation: newBaseOperation(
			model2.TypeOfOperation_DOC_ARR_DEL,
			nil,
			&DocDeleteInArrayBody{
				P: parent,
			},
		),
		Pos:        pos,
		NumOfNodes: numOfNodes,
	}
}

type DocDeleteInArrayBody struct {
	P *model2.Timestamp
	T []*model2.Timestamp
}

// DocDeleteInArrayOperation is used to delete a value into JSONArray.
type DocDeleteInArrayOperation struct {
	baseOperation
	Pos        int // for local
	NumOfNodes int // for local
}

func (its *DocDeleteInArrayOperation) GetBody() *DocDeleteInArrayBody {
	return its.Body.(*DocDeleteInArrayBody)
}
