package datatypes

import (
	"github.com/knowhunger/ortoo/ortoo/model"
	"github.com/knowhunger/ortoo/ortoo/types"
)

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
func (d *DummyWire) OnChangeDatatypeState(dt types.Datatype, state model.StateOfDatatype) error {
	return nil
}
