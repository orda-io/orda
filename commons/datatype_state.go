package commons

type DatatypeState uint8

const (
	LocallyExisted DatatypeState = 0 + iota
	Registering
	Registered
	Unregistered
)
