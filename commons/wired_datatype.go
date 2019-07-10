package commons

type WiredDatatypeT struct {
	wire
	*BaseDatatypeT
	checkPoint *CheckPoint
	super      interface{}
}

type WiredDatatype interface {
	getBase() *BaseDatatypeT
	executeRemote(op Operation)
}

func newWiredDataType(t DatatypeType, w wire) *WiredDatatypeT {
	return &WiredDatatypeT{
		BaseDatatypeT: newBaseDatatypeT(t),
		checkPoint:    newCheckPoint(),
		wire:          w,
	}
}

func execute(datatype interface{}, op Operation) (interface{}, error) {
	wired := getWiredDatatypeT(datatype)
	ret, err := executeLocalBase(wired.BaseDatatypeT, datatype, op)
	if err != nil {
		return ret, err
	}
	wired.deliverOperation(wired, op)
	return ret, nil
}

func (c *WiredDatatypeT) getBase() *BaseDatatypeT {
	return c.BaseDatatypeT
}

func (c *WiredDatatypeT) String() string {
	return c.BaseDatatypeT.String()
}

func (c *WiredDatatypeT) executeRemote(op Operation) {
	c.opID.syncLamport(op.GetOperationID().lamport)
	op.executeRemote(c.super)
}
