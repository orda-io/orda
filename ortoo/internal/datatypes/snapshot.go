package datatypes

import (
	"github.com/knowhunger/ortoo/ortoo/types"
)

// SnapshotDatatype defines the interface of the datatype that enables the snapshot.
type SnapshotDatatype interface {
	SetSnapshot(snapshot types.Snapshot)
	GetSnapshot() types.Snapshot
	GetAsJSON() (string, error)
}
