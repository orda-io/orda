package iface

type PublicSnapshotDatatype interface {
	GetAsJSON() interface{}
}

// SnapshotDatatype defines the interface of the datatype that enables the snapshot.
type SnapshotDatatype interface {
	PublicSnapshotDatatype
	SetSnapshot(snapshot Snapshot)
	GetSnapshot() Snapshot
	GetMetaAndSnapshot() ([]byte, Snapshot, error)         // @Real datatype
	SetMetaAndSnapshot(meta []byte, snapshot string) error // @Real datatype
}

// Snapshot defines the interface for snapshot used in a datatype.
type Snapshot interface {
	// json.Unmarshaler
	// json.Marshaler
	CloneSnapshot() Snapshot
	GetAsJSONCompatible() interface{}
}
