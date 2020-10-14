package iface

import "github.com/knowhunger/ortoo/pkg/errors"

// PublicSnapshotDatatype defines public operations related to snapshot
type PublicSnapshotDatatype interface {
	GetAsJSON() interface{}
}

// SnapshotDatatype defines the interfaces of the datatype that enables the snapshot.
type SnapshotDatatype interface {
	PublicSnapshotDatatype
	SetSnapshot(snapshot Snapshot)
	GetSnapshot() Snapshot
	GetMetaAndSnapshot() ([]byte, []byte, errors.OrtooError)       // @Real datatype
	SetMetaAndSnapshot(meta []byte, snap []byte) errors.OrtooError // @Real datatype
}

// Snapshot defines the interfaces for snapshot used in a datatype.
type Snapshot interface {
	CloneSnapshot() Snapshot
	GetAsJSONCompatible() interface{}
}
