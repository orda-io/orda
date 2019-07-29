package testonly

import (
	"github.com/knowhunger/ortoo/commons/internal/datatypes"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
)

type TestWire struct {
	wiredList []datatypes.WiredDatatype
}

func NewTestWire() *TestWire {
	return &TestWire{
		wiredList: make([]datatypes.WiredDatatype, 0),
	}
}

func (c *TestWire) DeliverOperation(wired datatypes.WiredDatatype, op model.Operationer) {
	for _, w := range c.wiredList {
		if wired.GetBase() != w.GetBase() {
			log.Logger.Info(wired, " => ", w)
			w.ExecuteRemote(op)
		}
	}

}

func (c *TestWire) SetDatatypes(datatypeList ...interface{}) {
	for _, v := range datatypeList {
		if opExecutor, ok := v.(datatypes.WiredDatatyper); ok {
			c.wiredList = append(c.wiredList, opExecutor.GetWired())
		}
	}
}
