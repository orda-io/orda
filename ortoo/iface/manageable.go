package iface

import (
	"github.com/knowhunger/ortoo/ortoo/errors"
	"github.com/knowhunger/ortoo/ortoo/model"
)

type ManageableDatatype interface {
	SubscribeOrCreate(state model.StateOfDatatype) errors.OrtooError                                             // @ManageableDatatype
	ExecuteRemoteTransaction(transaction []*model.Operation, obtainList bool) ([]interface{}, errors.OrtooError) // @ManageableDatatype
}
