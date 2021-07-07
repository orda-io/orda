package iface

// Snapshot defines the interfaces for snapshot used in a datatype. Snapshot contains metadata
type Snapshot interface {
	ToJSON() interface{}
}
