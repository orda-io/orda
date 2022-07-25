package iface

import (
	"github.com/orda-io/orda/client/pkg/errors"
	"github.com/orda-io/orda/client/pkg/model"
)

// Wire defines the interfaces related to delivering operations. This is called when a datatype needs to send messages
type Wire interface {
	DeliverTransaction(wired WiredDatatype)
	OnChangeDatatypeState(dt Datatype, state model.StateOfDatatype) errors.OrdaError
}
