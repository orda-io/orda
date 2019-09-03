package datatypes

import (
	"github.com/knowhunger/ortoo/commons/internal/constants"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
)

type WiredDatatypeImpl struct {
	Wire
	*baseDatatype
	trans      *WiredDatatypeImpl
	checkPoint *model.CheckPoint
	buffer     []*model.OperationOnWire
}

type WiredDatatyper interface {
	GetWired() WiredDatatype
}

type PublicWiredDatatypeInterface interface {
	PublicBaseDatatypeInterface
}

type WiredDatatype interface {
	GetBase() *baseDatatype
	ExecuteRemote(op model.Operation)
	ReceiveRemoteOperations(operations []model.Operation) error
	CreatePushPullPack() *model.PushPullPack
	ApplyPushPullPack(*model.PushPullPack)
}

func NewWiredDataType(t model.TypeDatatype, w Wire) (*WiredDatatypeImpl, error) {
	baseDatatype, err := newBaseDatatype(t)
	if err != nil {
		return nil, log.OrtooError(err, "fail to create wiredDatatype due to baseDatatype")
	}
	return &WiredDatatypeImpl{
		baseDatatype: baseDatatype,
		checkPoint:   model.NewCheckPoint(),
		buffer:       make([]*model.OperationOnWire, 0, constants.OperationBufferSize),
		Wire:         w,
	}, nil
}

func (w *WiredDatatypeImpl) GetBase() *baseDatatype {
	return w.baseDatatype
}

func (w *WiredDatatypeImpl) String() string {
	return w.baseDatatype.String()
}

func (w *WiredDatatypeImpl) ExecuteRemote(op model.Operation) {
	w.opID.SyncLamport(op.GetBase().GetId().Lamport)
	w.GetBase().executeRemoteBase(op)
}

func (w *WiredDatatypeImpl) ReceiveRemoteOperations(operations []model.Operation) error {
	i := 0
	transactionDatatype := w.opExecuter.(TransactionDatatype)

	for i < len(operations) {
		op := operations[i]
		var transaction []model.Operation
		switch cast := op.(type) {
		case *model.TransactionOperation:
			transaction = operations[i : i+int(cast.NumOfOps)]
			i += int(cast.NumOfOps)
		default:
			transaction = []model.Operation{op}
			i++
		}
		err := transactionDatatype.ExecuteTransactionRemote(transaction)
		if err != nil {
			return w.Logger.OrtooError(err, "fail to execute Transaction")
		}
	}
	return nil
}

func (w *WiredDatatypeImpl) CreatePushPullPack() *model.PushPullPack {
	seq := w.checkPoint.Cseq
	operations := w.getOperationOnWires(seq + 1)
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
		w.ExecuteRemote(model.ToOperation(op))
	}

}

func (w *WiredDatatypeImpl) getOperationOnWires(cseq uint64) []*model.OperationOnWire {
	op := model.ToOperation(w.buffer[0])
	startCseq := op.GetBase().Id.GetSeq()
	var start = int(cseq - startCseq)
	if len(w.buffer) > start {
		return w.buffer[start:]
	}
	return []*model.OperationOnWire{}
}

func (w *WiredDatatypeImpl) deliverTransaction(transaction []model.Operation) {
	for _, op := range transaction {
		w.buffer = append(w.buffer, model.ToOperationOnWire(op))
	}
	w.DeliverTransaction(w, transaction)
}