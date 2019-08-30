package datatypes

import (
	"github.com/knowhunger/ortoo/commons/model"
)

type DummyWire struct {
}

func NewDummyWire() *DummyWire {
	return &DummyWire{}
}

func (d *DummyWire) DeliverOperation(wired WiredDatatype, ops model.Operation) {
}

func (d *DummyWire) DeliverTransaction(wired WiredDatatype, transaction []model.Operation) {
}
