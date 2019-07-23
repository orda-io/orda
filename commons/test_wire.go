package commons

import (
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
)

type TestWire struct {
	wiredList []WiredDatatype
}

func NewTestWire() *TestWire {
	return &TestWire{
		wiredList: make([]WiredDatatype, 0),
	}
}

func (c *TestWire) deliverOperation(wired WiredDatatype, op model.Operationer) {
	for _, w := range c.wiredList {
		if w.getBase() != wired.getBase() {
			log.Logger.Info(wired, " => ", w)
			w.executeRemote(op)
		}
	}

}

func (c *TestWire) SetDatatypes(datatypes ...WiredDatatype) {
	c.wiredList = append(c.wiredList, datatypes...)
}
