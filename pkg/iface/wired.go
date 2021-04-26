package iface

import (
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/model"
)

// Wire defines the interfaces related to delivering operations. This is called when a datatype needs to send messages
type Wire interface {
	DeliverTransaction(wired WiredDatatype)
	OnChangeDatatypeState(dt Datatype, state model.StateOfDatatype) errors.OrtooError
}

// WiredDatatype defines the internal interface related to the synchronization with Ortoo server
type WiredDatatype interface {
	BaseDatatype
	ResetWired()
	ReceiveRemoteModelOperations(ops []*model.Operation, obtainList bool) ([]interface{}, errors.OrtooError)
	ApplyPushPullPack(*model.PushPullPack)
	CreatePushPullPack() *model.PushPullPack
	NeedPull(sseq uint64) bool
	NeedPush() bool
}
