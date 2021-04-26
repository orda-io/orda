package iface

import (
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/model"
)

// Operation defines the interfaces of any operation
type Operation interface {
	SetID(opID *model.OperationID)
	ToModelOperation() *model.Operation
	GetType() model.TypeOfOperation
	String() string
	GetID() *model.OperationID
	GetAsJSON() interface{}
}

// OperationalDatatype defines interfaces related to executing operations.
type OperationalDatatype interface {
	ExecuteLocal(op interface{}) (interface{}, errors.OrtooError)  // @Real datatype
	ExecuteRemote(op interface{}) (interface{}, errors.OrtooError) // @Real datatype
}
