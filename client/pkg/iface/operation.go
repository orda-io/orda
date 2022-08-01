package iface

import (
	"github.com/orda-io/orda/client/pkg/model"
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
