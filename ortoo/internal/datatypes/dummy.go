package datatypes

import "github.com/knowhunger/ortoo/ortoo/model"

// DummyWire ...
type DummyWire struct {
}

// NewDummyWire ...
func NewDummyWire() *DummyWire {
	return &DummyWire{}
}

// DeliverTransaction ...
func (d *DummyWire) DeliverTransaction(wired *WiredDatatype) {
}

// OnChangeDatatypeState ...
func (d *DummyWire) OnChangeDatatypeState(dt model.Datatype, state model.StateOfDatatype) error {
	return nil
}
