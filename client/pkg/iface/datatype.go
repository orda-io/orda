package iface

import (
	"github.com/orda-io/orda/client/pkg/errors"
	"github.com/orda-io/orda/client/pkg/log"
	model2 "github.com/orda-io/orda/client/pkg/model"
)

// BaseDatatype defines a base operations for datatype
type BaseDatatype interface {
	GetType() model2.TypeOfDatatype
	GetState() model2.StateOfDatatype
	GetKey() string                        // @baseDatatype
	GetDUID() string                       // @baseDatatype
	SetDUID(duid string)                   // @baseDatatype
	GetCUID() string                       // @baseDatatype
	SetState(state model2.StateOfDatatype) // @baseDatatype
	SetLogger(l *log.OrdaLog)              // @baseDatatype
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
	ReceiveRemoteModelOperations(ops []*model2.Operation, obtainList bool) ([]interface{}, errors.OrdaError)
	ApplyPushPullPack(*model2.PushPullPack)
	CreatePushPullPack() *model2.PushPullPack
	DeliverTransaction(transaction []Operation)
	NeedPull(sseq uint64) bool
	NeedPush() bool
	SubscribeOrCreate(state model2.StateOfDatatype) errors.OrdaError
	ResetWired()
}

// OperationalDatatype defines interfaces related to executing operations.
type OperationalDatatype interface {
	ExecuteLocal(op interface{}) (interface{}, errors.OrdaError)  // @Real datatype
	ExecuteRemote(op interface{}) (interface{}, errors.OrdaError) // @Real datatype
	ExecuteRemoteTransaction(transaction []*model2.Operation, obtainList bool) ([]interface{}, errors.OrdaError)
}

// HandleableDatatype defines handlers for Orda datatype
type HandleableDatatype interface {
	HandleStateChange(oldState, newState model2.StateOfDatatype)
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
