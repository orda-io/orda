package iface

import (
	"github.com/knowhunger/ortoo/pkg/model"
)

// Operation defines the interfaces of any operation
type Operation interface {
	GetType() model.TypeOfOperation
	String() string
	GetID() *model.OperationID
	SetID(opID *model.OperationID)
	ToJSON() interface{}
	ToModelOperation() *model.Operation
}
