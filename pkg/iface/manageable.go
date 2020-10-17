package iface

import (
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/model"
)

// ManageableDatatype defines interfaces related to managing datatype
type ManageableDatatype interface {
	SubscribeOrCreate(state model.StateOfDatatype) errors.OrtooError                                             // @ManageableDatatype
	ExecuteRemoteTransaction(transaction []*model.Operation, obtainList bool) ([]interface{}, errors.OrtooError) // @ManageableDatatype
}
