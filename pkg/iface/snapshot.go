package iface

import (
	"github.com/knowhunger/ortoo/pkg/errors"
)

// PublicSnapshotDatatype defines public operations related to snapshot
type PublicSnapshotDatatype interface {
	GetAsJSON() interface{}
}

type SnapshotContent interface {
	GetS() string
}

// SnapshotDatatype defines the interfaces of the datatype that enables the snapshot.
type SnapshotDatatype interface {
	PublicSnapshotDatatype
	SetSnapshot(snapshot Snapshot)
	GetSnapshot() Snapshot
	ResetSnapshot()
	GetMetaAndSnapshot() ([]byte, []byte, errors.OrtooError)
	SetMetaAndSnapshot(meta []byte, snap []byte) errors.OrtooError
	ApplySnapshotOperation(sc SnapshotContent, newSnap Snapshot) errors.OrtooError
}

// Snapshot defines the interfaces for snapshot used in a datatype.
type Snapshot interface {
	SetBase(base BaseDatatype)
	GetBase() BaseDatatype
	GetAsJSONCompatible() interface{}
}
