package commons

type wire interface {
	deliverOperation(wired WiredDatatype, op Operation)
}

type defaultWire struct {
}
