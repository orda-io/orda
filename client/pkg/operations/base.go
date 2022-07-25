package operations

import (
	"fmt"
	model2 "github.com/orda-io/orda/client/pkg/model"
)

// ////////////////// baseOperation ////////////////////

func newBaseOperation(typeOf model2.TypeOfOperation, opID *model2.OperationID, content interface{}) baseOperation {
	return baseOperation{
		Type: typeOf,
		ID:   opID,
		Body: content,
	}
}

type baseOperation struct {
	Type model2.TypeOfOperation
	ID   *model2.OperationID
	Body interface{}
}

func (its *baseOperation) SetID(opID *model2.OperationID) {
	its.ID = opID
}

func (its *baseOperation) GetID() *model2.OperationID {
	return its.ID
}

func (its *baseOperation) GetTimestamp() *model2.Timestamp {
	return its.ID.GetTimestamp()
}

func (its *baseOperation) GetType() model2.TypeOfOperation {
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
	body := its.Body
	switch body.(type) {
	case []byte:
		body = string(its.Body.([]byte))
	}
	return fmt.Sprintf("%s(%s|%+v)", its.Type, its.ID.ToString(), body)
}

func (its *baseOperation) ToModelOperation() *model2.Operation {
	return &model2.Operation{
		ID:     its.ID,
		OpType: its.Type,
		Body:   marshalBody(its.Body),
	}
}
