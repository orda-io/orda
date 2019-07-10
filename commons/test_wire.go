package commons

import "github.com/knowhunger/ortoo/commons/utils"

type TestWire struct {
	wiredList []WiredDatatype
}

func NewTestWire() *TestWire {
	return &TestWire{
		wiredList: make([]WiredDatatype, 0),
	}
}

func (c *TestWire) deliverOperation(wired WiredDatatype, op Operation) {
	for _, w := range c.wiredList {
		if w.getBase() != wired.getBase() {
			utils.Log.Info(wired, "=>", w)
			w.executeRemote(op)
		}
	}

}

func (c *TestWire) SetDatatypes(datatypes ...WiredDatatype) {
	c.wiredList = append(c.wiredList, datatypes...)
}
