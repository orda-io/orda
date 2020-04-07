package model

// Snapshot defines the interface for snapshot used in a datatype.
type Snapshot interface {
	CloneSnapshot() Snapshot
	GetAsJSON() interface{}
}
