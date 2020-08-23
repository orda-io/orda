package iface

import "github.com/knowhunger/ortoo/ortoo/errors"

type PublicSnapshotDatatype interface {
	GetAsJSON() interface{}
}

// SnapshotDatatype defines the interface of the datatype that enables the snapshot.
type SnapshotDatatype interface {
	PublicSnapshotDatatype
	SetSnapshot(snapshot Snapshot)
	GetSnapshot() Snapshot
	GetMetaAndSnapshot() ([]byte, Snapshot, errors.OrtooError)         // @Real datatype
	SetMetaAndSnapshot(meta []byte, snapshot string) errors.OrtooError // @Real datatype
}

// Snapshot defines the interface for snapshot used in a datatype.
type Snapshot interface {
	CloneSnapshot() Snapshot
	GetAsJSONCompatible() interface{}
}
