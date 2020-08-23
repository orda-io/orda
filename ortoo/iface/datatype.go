package iface

import (
	"github.com/knowhunger/ortoo/ortoo/errors"
	"github.com/knowhunger/ortoo/ortoo/model"
)

// Datatype defines the interface of executing operations, which is implemented by every datatype.
type Datatype interface {
	WiredDatatype
	SnapshotDatatype
	ManageableDatatype
	OperationalDatatype
	Handler
}

type Handler interface {
	HandleStateChange(oldState, newState model.StateOfDatatype)
	HandleErrors(err ...errors.OrtooError)
	HandleRemoteOperations(operations []interface{})
}
