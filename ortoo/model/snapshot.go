package model

import "github.com/gogo/protobuf/types"

// Snapshot defines the interface for snapshot used in a datatype.
type Snapshot interface {
	CloneSnapshot() Snapshot
	GetTypeURL() string
	GetTypeAny() (*types.Any, error)
	GetAsJSON() (string, error)
}
