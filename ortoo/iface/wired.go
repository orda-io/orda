package iface

import (
	"github.com/knowhunger/ortoo/ortoo/errors"
	"github.com/knowhunger/ortoo/ortoo/model"
)

// Wire defines the interfaces related to delivering operations. This is called when a datatype needs to send messages
type Wire interface {
	DeliverTransaction(wired WiredDatatype)
	OnChangeDatatypeState(dt Datatype, state model.StateOfDatatype) errors.OrtooError
}

// WiredDatatype defines the internal interface related to the synchronization with Ortoo server
type WiredDatatype interface {
	BaseDatatype
	ReceiveRemoteModelOperations(ops []*model.Operation) ([]interface{}, errors.OrtooError)
	ApplyPushPullPack(*model.PushPullPack)
	CreatePushPullPack() *model.PushPullPack
	NeedSync(sseq uint64) bool
}
