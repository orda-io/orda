package datatypes

import (
	"github.com/knowhunger/ortoo/commons/internal/constants"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
)

type WiredDatatypeImpl struct {
	Wire
	*transactionalDatatypeImpl
	checkPoint *model.CheckPoint
	buffer     []*model.Operation
	opExecuter model.OperationExecuter
}

type WiredDatatyper interface {
	GetWired() WiredDatatype
}

type PublicWireInterface interface {
	PublicTransactionalInterface
}

type WiredDatatype interface {
	GetBase() *baseDatatypeImpl
	GetTransactional() *transactionalDatatypeImpl
	ExecuteRemote(op model.Operationer)
	CreatePushPullPack() *model.PushPullPack
	ApplyPushPullPack(*model.PushPullPack)
}

func NewWiredDataType(t model.TypeDatatype, w Wire) (*WiredDatatypeImpl, error) {
	transactionalDatatype, err := newTransactionalDatatype(t)
	if err != nil {
		return nil, log.OrtooError(err, "fail to create wiredDatatype due to baseDatatypeImpl")
	}
	return &WiredDatatypeImpl{
		transactionalDatatypeImpl: transactionalDatatype,
		checkPoint:                model.NewCheckPoint(),
		buffer:                    make([]*model.Operation, 0, constants.OperationBufferSize),
		Wire:                      w,
	}, nil
}

func (w *WiredDatatypeImpl) ExecuteWired(op model.Operationer) (interface{}, error) {
	ret, err := w.executeLocalTransactional(w.opExecuter, op)
	if err != nil {
		return ret, log.OrtooError(err, "fail to executeLocalTransactional")
	}

	w.buffer = append(w.buffer, model.ToOperation(op))
	w.DeliverOperation(w, op)
	return ret, nil
}

func (w *WiredDatatypeImpl) GetBase() *baseDatatypeImpl {
	return w.baseDatatypeImpl
}

func (w *WiredDatatypeImpl) GetTransactional() *transactionalDatatypeImpl {
	return w.transactionalDatatypeImpl
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

func (w *WiredDatatypeImpl) BeginTransactionOnWired() error {
	op, err := w.BeginTransaction()
	if err != nil {
		return log.OrtooError(err, "fail to begin transaction on wired")
	}
	w.buffer = append(w.buffer, model.ToOperation(op))
	return nil
}

func (w *WiredDatatypeImpl) EndTransactionOnWired() {
	op := w.EndTransaction()
	w.buffer = append(w.buffer, model.ToOperation(op))
}
