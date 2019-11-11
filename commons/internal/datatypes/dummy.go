package datatypes

//DummyWire ...
type DummyWire struct {
}

//NewDummyWire ...
func NewDummyWire() *DummyWire {
	return &DummyWire{}
}

////DeliverOperation ...
//func (d *DummyWire) DeliverOperation(wired WiredDatatypeInterface, ops model.Operation) {
//}

//DeliverTransaction ...
func (d *DummyWire) DeliverTransaction(wired *WiredDatatype) { //, transaction []model.Operation) {
}
