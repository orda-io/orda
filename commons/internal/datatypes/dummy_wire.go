package datatypes

import (
	"github.com/knowhunger/ortoo/commons/model"
)

//DummyWire ...
type DummyWire struct {
}

//NewDummyWire ...
func NewDummyWire() *DummyWire {
	return &DummyWire{}
}

//DeliverOperation ...
func (d *DummyWire) DeliverOperation(wired WiredDatatype, ops model.Operation) {
}

//DeliverTransaction ...
func (d *DummyWire) DeliverTransaction(wired WiredDatatype, transaction []model.Operation) {
}
