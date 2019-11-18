package datatypes

import "github.com/knowhunger/ortoo/commons/model"

//SnapshotDatatype defines the interface of the datatype that enables the snapshot.
type SnapshotDatatype interface {
	SetSnapshot(snapshot model.Snapshot)
	GetSnapshot() model.Snapshot
}
