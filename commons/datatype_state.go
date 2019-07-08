package commons

type DatatypeState uint8

const (
	StateLocallyExisted DatatypeState = 0 + iota
	StateRegistering
	StateRegistered
	StateUnregistered
)
