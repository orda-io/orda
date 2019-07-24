package datatypes

import (
	"github.com/knowhunger/ortoo/commons/internal/constants"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
)

type WiredDatatypeImpl struct {
	Wire
	*baseDatatype
	checkPoint *model.CheckPoint
	buffer     []model.Operationer
	opExecuter model.OperationExecuter
}

type WiredDatatype interface {
	GetBase() *baseDatatype
	ExecuteRemote(op model.Operationer)
}

func NewWiredDataType(t model.TypeDatatype, w Wire) (*WiredDatatypeImpl, error) {
	baseDatatype, err := newBaseDatatype(t)
	if err != nil {
		return nil, log.OrtooError(err, "fail to create wiredDatatype due to baseDatatype")
	}
	return &WiredDatatypeImpl{
		baseDatatype: baseDatatype,
		checkPoint:   model.NewCheckPoint(),
		buffer:       make([]model.Operationer, constants.OperationBufferSize),
		Wire:         w,
	}, nil
}

func (w *WiredDatatypeImpl) ExecuteWired(datatype model.OperationExecuter, op model.Operationer) (interface{}, error) {
	//wired := commons.getWiredDatatypeT(datatype)
	ret, err := w.executeBase(datatype, op)
	if err != nil {
		return ret, err
	}
	w.buffer = append(w.buffer, op)
	w.DeliverOperation(w, op)
	return ret, nil
}

func (w *WiredDatatypeImpl) GetBase() *baseDatatype {
	return w.baseDatatype
}

func (w *WiredDatatypeImpl) String() string {
	return w.baseDatatype.String()
}

func (w *WiredDatatypeImpl) SetOperationExecuter(opExecuter model.OperationExecuter) {
	w.opExecuter = opExecuter
}

func (w *WiredDatatypeImpl) ExecuteRemote(op model.Operationer) {
	w.opID.SyncLamport(op.GetBase().GetId().Lamport)
	op.ExecuteRemote(w.opExecuter)
}

func (w *WiredDatatypeImpl) createPushPullPack() {
	seq := w.checkPoint.Cseq
	operations := w.getOperations(seq + 1)
	cp := &model.CheckPoint{}
	cp.Set(w.checkPoint.GetSseq(), w.checkPoint.GetCseq()+uint64(len(operations)))

}

func (w *WiredDatatypeImpl) getOperations(cseq uint64) []model.Operationer {
	startCseq := w.buffer[0].GetBase().GetId().Seq
	var start = int(cseq - uint64(startCseq))
	if len(w.buffer) > start {
		return w.buffer[start:]
	}
	return []model.Operationer{}

}
