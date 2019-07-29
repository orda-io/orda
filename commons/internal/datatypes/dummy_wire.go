package datatypes

import "github.com/knowhunger/ortoo/commons/model"

type DummyWire struct {
}

func NewDummyWire() *DummyWire {
	return &DummyWire{}
}

func (d *DummyWire) DeliverOperation(wired WiredDatatype, op model.Operationer) {

}
