package iface

import (
	model2 "github.com/orda-io/orda/client/pkg/model"
)

// Operation defines the interfaces of any operation
type Operation interface {
	GetType() model2.TypeOfOperation
	String() string
	GetID() *model2.OperationID
	SetID(opID *model2.OperationID)
	ToJSON() interface{}
	ToModelOperation() *model2.Operation
}
