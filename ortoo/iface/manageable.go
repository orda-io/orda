package iface

import "github.com/knowhunger/ortoo/ortoo/model"

type ManageableDatatype interface {
	SubscribeOrCreate(state model.StateOfDatatype) error                                             // @ManageableDatatype
	ExecuteTransactionRemote(transaction []*model.Operation, obtainList bool) ([]interface{}, error) // @ManageableDatatype
}
