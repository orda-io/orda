package testonly

import (
	"github.com/knowhunger/ortoo/ortoo/internal/datatypes"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/ortoo/model"
)

// TestWire ...
type TestWire struct {
	datatypeList []*datatypes.FinalDatatype
	sseqMap      map[string]int
}

// NewTestWire ...
func NewTestWire() *TestWire {
	return &TestWire{
		datatypeList: make([]*datatypes.FinalDatatype, 0),
		sseqMap:      make(map[string]int),
	}
}

// DeliverTransaction ...
func (c *TestWire) DeliverTransaction(wired *datatypes.WiredDatatype) {
	pushPullPack := wired.CreatePushPullPack()
	sseq := c.sseqMap[wired.GetBase().GetCUID()]
	operations := pushPullPack.Operations[sseq:]
	log.Logger.Infof("deliver transaction:%v", operations)
	c.sseqMap[wired.GetBase().GetCUID()] = len(pushPullPack.Operations)
	for _, w := range c.datatypeList {
		if wired != w.GetWired() {
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
func (c *TestWire) SetDatatypes(datatypeList ...*datatypes.FinalDatatype) {

	for _, v := range datatypeList {
		c.datatypeList = append(c.datatypeList, v)
		c.sseqMap[v.GetCUID()] = 0
	}
}
