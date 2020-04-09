package testonly

import (
	"github.com/knowhunger/ortoo/ortoo/iface"
	"github.com/knowhunger/ortoo/ortoo/internal/datatypes"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/ortoo/model"
)

// TestWire ...
type TestWire struct {
	datatypeList  []*datatypes.ManageableDatatype
	sseqMap       map[string]int
	deliveredList []iface.WiredDatatype
	notifiable    bool
}

// NewTestWire ...
func NewTestWire(notifiable bool) *TestWire {
	return &TestWire{
		datatypeList: make([]*datatypes.ManageableDatatype, 0),
		sseqMap:      make(map[string]int),
		notifiable:   notifiable,
	}
}

// DeliverTransaction ...
func (its *TestWire) DeliverTransaction(wired iface.WiredDatatype) {
	its.deliveredList = append(its.deliveredList, wired)
	if its.notifiable {
		its.Sync()
	}
}

// OnChangeDatatypeState ...
func (its *TestWire) OnChangeDatatypeState(dt iface.Datatype, state model.StateOfDatatype) error {
	return nil
}

// SetDatatypes ...
func (its *TestWire) SetDatatypes(datatypeList ...*datatypes.ManageableDatatype) {
	for _, v := range datatypeList {
		its.datatypeList = append(its.datatypeList, v)
		its.sseqMap[v.GetCUID()] = 0
	}
}

func (its *TestWire) Sync() {
	for len(its.deliveredList) > 0 {
		var wired iface.WiredDatatype
		wired, its.deliveredList = its.deliveredList[0], its.deliveredList[1:]

		pushPullPack := wired.CreatePushPullPack()
		sseq := its.sseqMap[wired.GetCUID()]
		operations := pushPullPack.Operations[sseq:]

		log.Logger.Infof("deliver transaction:%v", model.OperationsToString(operations))
		its.sseqMap[wired.GetCUID()] = len(pushPullPack.Operations)
		for _, w := range its.datatypeList {
			if wired != w.GetWired() {
				log.Logger.Info(wired.GetCUID(), " => ", w.GetCUID())
				w.ReceiveRemoteModelOperations(operations)
			}
		}
	}
}
