package datatypes

import (
	"github.com/knowhunger/ortoo/commons/model"
)

type Wire interface {
	DeliverOperation(wired WiredDatatype, op model.Operationer)
}

type defaultWire struct {
}
