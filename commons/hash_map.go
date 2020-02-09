package commons

import "github.com/knowhunger/ortoo/commons/internal/datatypes"

type HashMap interface {
	datatypes.PublicWiredDatatypeInterface
}

type HashMapInTxn interface {
	Get(key string) interface{}
	Put(key string, value interface{}) interface{}
}
