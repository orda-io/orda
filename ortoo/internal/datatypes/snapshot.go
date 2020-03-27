package datatypes

import "github.com/knowhunger/ortoo/ortoo/model"

// SnapshotDatatype defines the interface of the datatype that enables the snapshot.
type SnapshotDatatype interface {
	SetSnapshot(snapshot model.Snapshot)
	GetSnapshot() model.Snapshot
	GetAsJSON() (string, error)
}
