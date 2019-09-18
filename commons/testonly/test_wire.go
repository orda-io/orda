package testonly

import (
	"github.com/knowhunger/ortoo/commons/internal/datatypes"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
)

//TestWire ...
type TestWire struct {
	wiredList []datatypes.WiredDatatype
	list      []model.Operation
}

//NewTestWire ...
func NewTestWire() *TestWire {
	return &TestWire{
		wiredList: make([]datatypes.WiredDatatype, 0),
	}
}

//DeliverOperation ...
func (c *TestWire) DeliverOperation(wired datatypes.WiredDatatype, op model.Operation) {
	for _, w := range c.wiredList {
		if wired != w {
			log.Logger.Info(wired, " => ", w)
			w.ExecuteRemote(op)
		}
	}
	c.list = append(c.list, op)
}

//DeliverTransaction ...
func (c *TestWire) DeliverTransaction(wired datatypes.WiredDatatype, transaction []model.Operation) {
	for _, w := range c.wiredList {
		if wired != w {
			log.Logger.Info(wired, " => ", w)
			w.ReceiveRemoteOperations(transaction)
		}
	}
	c.list = append(c.list, transaction...)
}

//SetDatatypes ...
func (c *TestWire) SetDatatypes(datatypeList ...interface{}) {
	for _, v := range datatypeList {
		if opExecutor, ok := v.(datatypes.WiredDatatyper); ok {
			c.wiredList = append(c.wiredList, opExecutor.GetWired())
		}
	}
}
