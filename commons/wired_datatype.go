package commons

import (
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
)

type WiredDatatypeT struct {
	wire
	*baseDatatypeT
	checkPoint *model.CheckPoint
	buffer     []model.Operationer
	super      model.OperationExecuter
}

type WiredDatatype interface {
	getBase() *baseDatatypeT
	executeRemote(op model.Operationer)
}

func newWiredDataType(t DatatypeType, w wire) (*WiredDatatypeT, error) {
	baseDatatype, err := newBaseDatatypeT(t)
	if err != nil {
		return nil, log.OrtooError(err, "fail to create wiredDatatype due to baseDatatype")
	}
	return &WiredDatatypeT{
		baseDatatypeT: baseDatatype,
		checkPoint:    model.NewCheckPoint(),
		buffer:        make([]model.Operationer, operationBufferSize),
		wire:          w,
	}, nil
}

func (w *WiredDatatypeT) executeWired(datatype model.OperationExecuter, op model.Operationer) (interface{}, error) {
	wired := getWiredDatatypeT(datatype)
	ret, err := wired.executeBase(datatype, op)
	if err != nil {
		return ret, err
	}
	w.buffer = append(w.buffer, op)
	wired.deliverOperation(wired, op)
	return ret, nil
}

func (w *WiredDatatypeT) getBase() *baseDatatypeT {
	return w.baseDatatypeT
}

func (w *WiredDatatypeT) String() string {
	return w.baseDatatypeT.String()
}

func (w *WiredDatatypeT) executeRemote(op model.Operationer) {
	w.opID.SyncLamport(op.GetBase().GetOperationID().Lamport)
	op.ExecuteRemote(w.super)
}

func (w *WiredDatatypeT) createPushPullPack() {
	seq := w.checkPoint.Cseq
	operations := w.getOperations(seq + 1)
	cp := &model.CheckPoint{}
	cp.Set(w.checkPoint.GetSseq(), w.checkPoint.GetCseq()+uint64(len(operations)))

}

func (w *WiredDatatypeT) getOperations(cseq uint64) []model.Operationer {
	startCseq := w.buffer[0].GetBase().GetOperationID().Seq
	var start = int(cseq - uint64(startCseq))
	if len(w.buffer) > start {
		return w.buffer[start:]
	}
	return []model.Operationer{}

}
