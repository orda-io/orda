package datatypes

import (
	"github.com/knowhunger/ortoo/commons/internal/constants"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
)

type WiredDatatypeImpl struct {
	Wire
	*baseDatatypeImpl
	checkPoint *model.CheckPoint
	buffer     []*model.Operation
	opExecuter model.OperationExecuter
}

type WiredDatatyper interface {
	GetWired() WiredDatatype
}

type PublicWireInterface interface {
	PublicTransactionalInterface
	CreatePushPullPack() *model.PushPullPack
	ApplyPushPullPack(*model.PushPullPack)
}

type WiredDatatype interface {
	GetBase() *baseDatatypeImpl
	ExecuteRemote(op model.Operationer)
}

func NewWiredDataType(t model.TypeDatatype, w Wire) (*WiredDatatypeImpl, error) {
	baseDatatype, err := newBaseDatatype(t)
	if err != nil {
		return nil, log.OrtooError(err, "fail to create wiredDatatype due to baseDatatypeImpl")
	}
	return &WiredDatatypeImpl{
		baseDatatypeImpl: baseDatatype,
		checkPoint:       model.NewCheckPoint(),
		buffer:           make([]*model.Operation, 0, constants.OperationBufferSize),
		Wire:             w,
	}, nil
}

func (w *WiredDatatypeImpl) ExecuteWired(op model.Operationer) (interface{}, error) {
	ret, err := w.executeLocalBase(w.opExecuter, op)
	if err != nil {
		return ret, err
	}

	w.buffer = append(w.buffer, model.ToOperation(op))
	w.DeliverOperation(w, op)
	return ret, nil
}

func (w *WiredDatatypeImpl) GetBase() *baseDatatypeImpl {
	return w.baseDatatypeImpl
}

func (w *WiredDatatypeImpl) String() string {
	return w.baseDatatypeImpl.String()
}

func (w *WiredDatatypeImpl) SetOperationExecuter(opExecuter model.OperationExecuter) {
	w.opExecuter = opExecuter
}

func (w *WiredDatatypeImpl) ExecuteRemote(op model.Operationer) {
	w.opID.SyncLamport(op.GetBase().GetId().Lamport)
	w.GetBase().executeRemoteBase(w.opExecuter, op)
	//op.ExecuteRemote(w.opExecuter)
}

func (w *WiredDatatypeImpl) CreatePushPullPack() *model.PushPullPack {
	seq := w.checkPoint.Cseq
	operations := w.getOperations(seq + 1)
	cp := &model.CheckPoint{
		Sseq: w.checkPoint.GetSseq(),
		Cseq: w.checkPoint.GetCseq() + uint64(len(operations)),
	}
	return &model.PushPullPack{
		CheckPoint: cp,
		Duid:       w.id,
		Era:        0,
		Type:       0,
		Operations: operations,
	}
}

func (w *WiredDatatypeImpl) ApplyPushPullPack(ppp *model.PushPullPack) {

	opList := ppp.GetOperations()
	for _, op := range opList {
		w.ExecuteRemote(model.ToOperationer(op))
	}

}

func (w *WiredDatatypeImpl) getOperations(cseq uint64) []*model.Operation {
	op := model.ToOperationer(w.buffer[0])
	startCseq := op.GetBase().Id.GetSeq()
	var start = int(cseq - uint64(startCseq))
	if len(w.buffer) > start {
		return w.buffer[start:]
	}
	return []*model.Operation{}

}
