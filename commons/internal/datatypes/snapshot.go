package datatypes

type Snapshot interface {
	CloneSnapshot() Snapshot
}

type SnapshotDatatype interface {
	SetSnapshot(snapshot Snapshot)
	GetSnapshot() Snapshot
}
