package iface

import "github.com/knowhunger/ortoo/ortoo/model"

// Operation defines the interfaces of Operation
type Operation interface {
	SetOperationID(opID *model.OperationID)
	ExecuteLocal(datatype Datatype) (interface{}, error)
	ExecuteRemote(datatype Datatype) (interface{}, error)
	ToModelOperation() *model.Operation
	GetType() model.TypeOfOperation
	String() string
	GetID() *model.OperationID
	GetAsJSON() interface{}
}

type OperationalDatatype interface {
	ExecuteLocal(op interface{}) (interface{}, error)  // @Real datatype
	ExecuteRemote(op interface{}) (interface{}, error) // @Real datatype
}
