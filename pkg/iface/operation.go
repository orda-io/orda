package iface

import (
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/model"
)

// Operation defines the interfaces of Operation
type Operation interface {
	SetOperationID(opID *model.OperationID)
	ExecuteLocal(datatype Datatype) (interface{}, errors.OrtooError)
	ExecuteRemote(datatype Datatype) (interface{}, errors.OrtooError)
	ToModelOperation() *model.Operation
	GetType() model.TypeOfOperation
	String() string
	GetID() *model.OperationID
	GetAsJSON() interface{}
}

type OperationalDatatype interface {
	ExecuteLocal(op interface{}) (interface{}, errors.OrtooError)  // @Real datatype
	ExecuteRemote(op interface{}) (interface{}, errors.OrtooError) // @Real datatype
}
