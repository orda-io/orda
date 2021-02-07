package iface

import (
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/model"
)

// Datatype defines the interface of executing operations, which is implemented by every datatype.
type Datatype interface {
	WiredDatatype
	ManageableDatatype
	OperationalDatatype
	SnapshotDatatype
	TransactionDatatype
	Handler
}

// Handler defines handlers for Ortoo datatype
type Handler interface {
	HandleStateChange(oldState, newState model.StateOfDatatype)
	HandleErrors(err ...errors.OrtooError)
	HandleRemoteOperations(operations []interface{})
}

/*
	iface.Datatype extends iface.WiredDatatype extends iface.BaseDatatype extends iface.PublicBaseDatatype
	iface.Datatype extends iface.ManageableDatatype
	iface.Datatype extends iface.OperationalDatatype
	iface.Datatype extends iface.SnapshotDatatype
	iface.Datatype extends iface.Handler

	ManageableDatatype extends TransactionDatatype extends datatypes.WiredDatatype extends datatypes.BaseDatatype
	datatypes.BaseDatatype implements iface.BaseDatatype
	datatypes.WiredDatatype implements iface.WiredDatatype
	datatypes.TransactionDatatype implements iface.WiredDatatype
*/
