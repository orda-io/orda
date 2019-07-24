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
		if wired.GetBase() != wired.GetBase() {
			log.Logger.Info(wired, " => ", w)
			wired.ExecuteRemote(op)
		}
	}

}

func (c *TestWire) SetDatatypes(operationExecuters ...interface{}) {
	var wiredDatatypes []datatypes.WiredDatatype

	for _, v := range operationExecuters {
		wiredDatatype, ok := v.(datatypes.WiredDatatype)
		if !ok {
			continue
		}
		wiredDatatypes = append(wiredDatatypes, wiredDatatype)
	}
	c.wiredList = append(c.wiredList, wiredDatatypes...)
}
