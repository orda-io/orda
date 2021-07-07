package operations

import (
	"fmt"
	"github.com/orda-io/orda/pkg/model"
)

// ////////////////// baseOperation ////////////////////

func newBaseOperation(typeOf model.TypeOfOperation, opID *model.OperationID, content interface{}) baseOperation {
	return baseOperation{
		Type: typeOf,
		ID:   opID,
		Body: content,
	}
}

type baseOperation struct {
	Type model.TypeOfOperation
	ID   *model.OperationID
	Body interface{}
}

func (its *baseOperation) SetID(opID *model.OperationID) {
	its.ID = opID
}

func (its *baseOperation) GetID() *model.OperationID {
	return its.ID
}

func (its *baseOperation) GetTimestamp() *model.Timestamp {
	return its.ID.GetTimestamp()
}

func (its *baseOperation) GetType() model.TypeOfOperation {
	return its.Type
}

// ToJSON returns the operation in the format of JSON compatible struct.
func (its *baseOperation) ToJSON() interface{} {
	return struct {
		ID   interface{}
		Type string
		Body interface{}
	}{
		ID:   its.ID.ToJSON(),
		Type: its.Type.String(),
		Body: its.Body,
	}
}

func (its *baseOperation) String() string {
	return fmt.Sprintf("%s(%s|%+v)", its.Type, its.ID.ToString(), its.Body)
}

func (its *baseOperation) ToModelOperation() *model.Operation {
	return &model.Operation{
		ID:     its.ID,
		OpType: its.Type,
		Body:   marshalBody(its.Body),
	}
}
