package commons

type DatatypeType uint8

const (
	TypeIntCounter DatatypeType = 1 + iota
	TypeJson
)
