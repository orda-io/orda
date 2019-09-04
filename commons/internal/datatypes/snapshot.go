package datatypes

//Snapshot defines the interface for snapshot used in a datatype.
type Snapshot interface {
	CloneSnapshot() Snapshot
}

//SnapshotDatatype defines the interface of the datatype that enables the snapshot.
type SnapshotDatatype interface {
	SetSnapshot(snapshot Snapshot)
	GetSnapshot() Snapshot
}
