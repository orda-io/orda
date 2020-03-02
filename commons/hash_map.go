package commons

import "github.com/knowhunger/ortoo/commons/internal/datatypes"

// HashMap is an Ortoo datatype which provides hash map interfaces.
type HashMap interface {
	datatypes.PublicWiredDatatypeInterface
}

// HashMapInTxn is an Ortoo datatype which provides hash map interface in a transaction.
type HashMapInTxn interface {
	Get(key string) interface{}
	Put(key string, value interface{}) interface{}
}
