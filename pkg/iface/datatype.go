package iface

import (
	"github.com/orda-io/orda/pkg/errors"
	"github.com/orda-io/orda/pkg/log"
	"github.com/orda-io/orda/pkg/model"
)

// BaseDatatype defines a base operations for datatype
type BaseDatatype interface {
	GetType() model.TypeOfDatatype
	GetState() model.StateOfDatatype
	GetKey() string                       // @baseDatatype
	GetDUID() string                      // @baseDatatype
	SetDUID(duid string)                  // @baseDatatype
	GetCUID() string                      // @baseDatatype
	SetState(state model.StateOfDatatype) // @baseDatatype
	SetLogger(l *log.OrdaLog)             // @baseDatatype
	GetMeta() ([]byte, errors.OrdaError)
	SetMeta(meta []byte) errors.OrdaError
	GetSummary() string
	L() *log.OrdaLog
}

// SnapshotDatatype defines the interfaces of the datatype that enables the snapshot.
type SnapshotDatatype interface {
	ResetSnapshot()
	GetSnapshot() Snapshot
	GetMetaAndSnapshot() ([]byte, []byte, errors.OrdaError)
	SetMetaAndSnapshot(meta []byte, snap []byte) errors.OrdaError
	CreateSnapshotOperation() (Operation, errors.OrdaError)
	ToJSON() interface{}
}

// WiredDatatype defines the internal interface related to the synchronization with Orda server
type WiredDatatype interface {
	BaseDatatype
	SetCheckPoint(sseq uint64, cseq uint64)
	ReceiveRemoteModelOperations(ops []*model.Operation, obtainList bool) ([]interface{}, errors.OrdaError)
	ApplyPushPullPack(*model.PushPullPack)
	CreatePushPullPack() *model.PushPullPack
	DeliverTransaction(transaction []Operation)
	NeedPull(sseq uint64) bool
	NeedPush() bool
	SubscribeOrCreate(state model.StateOfDatatype) errors.OrdaError
	ResetWired()
}

// OperationalDatatype defines interfaces related to executing operations.
type OperationalDatatype interface {
	ExecuteLocal(op interface{}) (interface{}, errors.OrdaError)  // @Real datatype
	ExecuteRemote(op interface{}) (interface{}, errors.OrdaError) // @Real datatype
	ExecuteRemoteTransaction(transaction []*model.Operation, obtainList bool) ([]interface{}, errors.OrdaError)
}

// HandleableDatatype defines handlers for Orda datatype
type HandleableDatatype interface {
	HandleStateChange(oldState, newState model.StateOfDatatype)
	HandleErrors(err ...errors.OrdaError)
	HandleRemoteOperations(operations []interface{})
}

// Datatype defines the interface of executing operations, which is implemented by every datatype.
type Datatype interface {
	SnapshotDatatype
	WiredDatatype
	OperationalDatatype
	HandleableDatatype
}
