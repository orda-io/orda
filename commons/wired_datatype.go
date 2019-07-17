package commons

type WiredDatatypeT struct {
	wire
	*BaseDatatypeT
	checkPoint *CheckPoint
	buffer     []Operation
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
		buffer:        make([]Operation, operationBufferSize),
		wire:          w,
	}
}

func (w *WiredDatatypeT) executeWired(datatype interface{}, op Operation) (interface{}, error) {
	wired := getWiredDatatypeT(datatype)
	ret, err := wired.executeBase(datatype, op)
	if err != nil {
		return ret, err
	}
	w.buffer = append(w.buffer, op)
	wired.deliverOperation(wired, op)
	return ret, nil
}

func (w *WiredDatatypeT) getBase() *BaseDatatypeT {
	return w.BaseDatatypeT
}

func (w *WiredDatatypeT) String() string {
	return w.BaseDatatypeT.String()
}

func (w *WiredDatatypeT) executeRemote(op Operation) {
	w.opID.syncLamport(op.GetOperationID().lamport)
	op.executeRemote(w.super)
}

func (w *WiredDatatypeT) createPushPullPack() {
	seq := w.checkPoint.Cseq
	operations := w.getOperations(seq + 1)
	cp := &CheckPoint{}
	cp.Set(w.checkPoint.GetSseq(), w.checkPoint.GetCseq()+uint64(len(operations)))

}

func (w *WiredDatatypeT) getOperations(cseq uint64) []Operation {
	startCseq := w.buffer[0].GetOperationID().seq
	var start = int(cseq - uint64(startCseq))
	if len(w.buffer) > start {
		return w.buffer[start:]
	}
	return []Operation{}

}
