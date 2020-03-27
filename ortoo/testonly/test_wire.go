package testonly

import (
	"github.com/knowhunger/ortoo/ortoo/internal/datatypes"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/ortoo/model"
)

// TestWire ...
type TestWire struct {
	wiredList []*datatypes.WiredDatatype //Interface
	sseqMap   map[string]int
}

// NewTestWire ...
func NewTestWire() *TestWire {
	return &TestWire{
		wiredList: make([]*datatypes.WiredDatatype, 0),
		sseqMap:   make(map[string]int),
	}
}

// DeliverTransaction ...
func (c *TestWire) DeliverTransaction(wired *datatypes.WiredDatatype) {
	pushPullPack := wired.CreatePushPullPack()
	sseq := c.sseqMap[wired.GetBase().GetCUID()]
	operations := pushPullPack.Operations[sseq:]
	c.sseqMap[wired.GetBase().GetCUID()] = len(pushPullPack.Operations)
	for _, w := range c.wiredList {
		if wired != w {
			log.Logger.Info(wired, " => ", w)
			w.ReceiveRemoteModelOperations(operations)
		}
	}
}

// OnChangeDatatypeState ...
func (c *TestWire) OnChangeDatatypeState(dt model.Datatype, state model.StateOfDatatype) error {
	return nil
}

// SetDatatypes ...
func (c *TestWire) SetDatatypes(datatypeList ...interface{}) {

	for _, v := range datatypeList {
		if cv, ok := v.(datatypes.FinalDatatypeInterface); ok {
			common := cv.GetFinal()
			c.wiredList = append(c.wiredList, common.GetWired())
			c.sseqMap[common.GetCUID()] = 0
		}
	}
}
