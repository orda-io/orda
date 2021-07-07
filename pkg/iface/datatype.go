package iface

import (
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/log"
	"github.com/knowhunger/ortoo/pkg/model"
)

// BaseDatatype defines a base operations for datatype
type BaseDatatype interface {
	GetType() model.TypeOfDatatype
	GetState() model.StateOfDatatype
	GetKey() string                       // @baseDatatype
	GetDUID() string                      // @baseDatatype
	GetCUID() string                      // @baseDatatype
	SetState(state model.StateOfDatatype) // @baseDatatype
	SetLogger(l *log.OrtooLog)            // @baseDatatype
	GetMeta() ([]byte, errors.OrtooError)
	SetMeta(meta []byte) errors.OrtooError
	GetSummary() string
	L() *log.OrtooLog
}

// SnapshotDatatype defines the interfaces of the datatype that enables the snapshot.
type SnapshotDatatype interface {
	ResetSnapshot()
	GetSnapshot() Snapshot
	GetMetaAndSnapshot() ([]byte, []byte, errors.OrtooError)
	SetMetaAndSnapshot(meta []byte, snap []byte) errors.OrtooError
	CreateSnapshotOperation() (Operation, errors.OrtooError)
	ToJSON() interface{}
}

// WiredDatatype defines the internal interface related to the synchronization with Ortoo server
type WiredDatatype interface {
	BaseDatatype
	ReceiveRemoteModelOperations(ops []*model.Operation, obtainList bool) ([]interface{}, errors.OrtooError)
	ApplyPushPullPack(*model.PushPullPack)
	CreatePushPullPack() *model.PushPullPack
	DeliverTransaction(transaction []Operation)
	NeedPull(sseq uint64) bool
	NeedPush() bool
	SubscribeOrCreate(state model.StateOfDatatype) errors.OrtooError
}

// OperationalDatatype defines interfaces related to executing operations.
type OperationalDatatype interface {
	ExecuteLocal(op interface{}) (interface{}, errors.OrtooError)  // @Real datatype
	ExecuteRemote(op interface{}) (interface{}, errors.OrtooError) // @Real datatype
	ExecuteRemoteTransaction(transaction []*model.Operation, obtainList bool) ([]interface{}, errors.OrtooError)
}

// HandleableDatatype defines handlers for Ortoo datatype
type HandleableDatatype interface {
	HandleStateChange(oldState, newState model.StateOfDatatype)
	HandleErrors(err ...errors.OrtooError)
	HandleRemoteOperations(operations []interface{})
}

// Datatype defines the interface of executing operations, which is implemented by every datatype.
type Datatype interface {
	SnapshotDatatype
	WiredDatatype
	OperationalDatatype
	HandleableDatatype
}
