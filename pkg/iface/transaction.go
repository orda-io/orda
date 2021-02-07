package iface

import "github.com/knowhunger/ortoo/pkg/errors"

type TransactionDatatype interface {
	ResetRollBackContext() errors.OrtooError
}
