package wrapper

import (
	"github.com/orda-io/orda/client/pkg/context"
	"github.com/orda-io/orda/client/pkg/iface"
	"github.com/orda-io/orda/client/pkg/model"
	"github.com/orda-io/orda/client/pkg/orda"
)

// DatatypeWrapper is used to manipulate datatype with internal APIs in orda server
type DatatypeWrapper struct {
	iface.Datatype
	reqNum uint32
}

// NewDatatypeWrapper creates a new DatatypeWrapper
func NewDatatypeWrapper(dt orda.Datatype) *DatatypeWrapper {
	return &DatatypeWrapper{
		Datatype: dt.(iface.Datatype),
		reqNum:   0,
	}
}

// GetClientModel returns model.Client from its context
func (its *DatatypeWrapper) GetClientModel() *model.Client {
	ctx := its.Datatype.(iface.BaseDatatype).GetCtx().(*context.DatatypeContext)
	return ctx.ClientContext.Client
}

// CreatePushPullPack creates PushPullPack
func (its *DatatypeWrapper) CreatePushPullPack() *model.PushPullPack {
	return its.Datatype.(iface.WiredDatatype).CreatePushPullPack()
}

// CreatePushPullMessage creates PushPullPackMessage
func (its *DatatypeWrapper) CreatePushPullMessage() *model.PushPullMessage {
	its.reqNum++
	return model.NewPushPullMessage(its.reqNum, its.GetClientModel(), its.CreatePushPullPack())
}

// ApplyPushPullPack applies PushPullPack
func (its *DatatypeWrapper) ApplyPushPullPack(ppp *model.PushPullPack) {
	its.Datatype.ApplyPushPullPack(ppp)
}
