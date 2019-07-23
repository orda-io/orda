package commons

import "github.com/knowhunger/ortoo/commons/model"

type wire interface {
	deliverOperation(wired WiredDatatype, op model.Operationer)
}

type defaultWire struct {
}
