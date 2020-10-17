package iface

import "github.com/knowhunger/ortoo/pkg/errors"

type TransactionDatabase interface {
	ResetRollBackContext() errors.OrtooError
}
