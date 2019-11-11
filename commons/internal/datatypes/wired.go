package datatypes

import (
	"github.com/knowhunger/ortoo/commons/internal/constants"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
)

// Wire defines the interfaces related to delivering operations. This is called when a dataType needs to send messages
type Wire interface {
	//DeliverOperation(wired WiredDatatypeInterface, op model.Operation)
	DeliverTransaction(wired *WiredDatatype)
}

// WiredDatatype implements the datatype features related to the synchronization with Ortoo server
type WiredDatatype struct {
	*baseDatatype
	wire       Wire
	checkPoint *model.CheckPoint
	buffer     []*model.OperationOnWire
}

//// WiredDatatyper defines the interface used in Wire
//type WiredDatatyper interface {
//	GetWired() WiredDatatypeInterface
//}

// PublicWiredDatatypeInterface defines the interface related to the synchronization with Ortoo server
type PublicWiredDatatypeInterface interface {
	PublicBaseDatatypeInterface
}

// WiredDatatypeInterface defines the internal interface related to the synchronization with Ortoo server
type WiredDatatypeInterface interface {
	GetBase() *baseDatatype
	ExecuteRemote(op model.Operation)
	ReceiveRemoteOperationsOnWire(operations []*model.OperationOnWire) error
	ApplyPushPullPack(*model.PushPullPack)
	CreatePushPullPack() *model.PushPullPack
}

// newWiredDataType creates a new wiredDatatype
func newWiredDataType(b *baseDatatype, w Wire) (*WiredDatatype, error) {
	return &WiredDatatype{
		baseDatatype: b,
		checkPoint:   model.NewCheckPoint(),
		buffer:       make([]*model.OperationOnWire, 0, constants.OperationBufferSize),
		wire:         w,
	}, nil
}

func (w *WiredDatatype) GetBase() *baseDatatype {
	return w.baseDatatype
}

func (w *WiredDatatype) String() string {
	return w.baseDatatype.String()
}

// ExecuteRemote ...
func (w *WiredDatatype) ExecuteRemote(op model.Operation) {
	w.opID.SyncLamport(op.GetBase().GetID().Lamport)
	w.executeRemoteBase(op)
}

// ReceiveRemoteOperationsOnWire ...
func (w *WiredDatatype) ReceiveRemoteOperationsOnWire(operations []*model.OperationOnWire) error {

	finalDatatype := w.finalDatatype

	for i := 0; i < len(operations); {
		op := model.ToOperation(operations[i])
		var transaction []model.Operation
		switch cast := op.(type) {
		case *model.TransactionOperation:
			transactionOnWire := operations[i : i+int(cast.NumOfOps)]
			for _, opOnWire := range transactionOnWire {
				transaction = append(transaction, model.ToOperation(opOnWire))
			}
			i += int(cast.NumOfOps)
		default:
			transaction = []model.Operation{op}
			i++
		}
		err := finalDatatype.ExecuteTransactionRemote(transaction)
		if err != nil {
			return w.Logger.OrtooErrorf(err, "fail to execute Transaction")
		}
	}
	return nil
}

func (w *WiredDatatype) applyPushPullPackExecuteOperations(operationsOnWire []*model.OperationOnWire) {
	w.ReceiveRemoteOperationsOnWire(operationsOnWire)
	//for i := 0; i < len(operationsOnWire); {
	//op := model.ToOperation(operationsOnWire[i])
	//var transaction []*model.Operation
	//switch cast := op.(type) {
	//case *model.TransactionOperation:

	//transaction = append(transaction, operationsOnWire[i : i+int(cast.NumOfOps)])

	//transaction = operationsOnWire[i : i+int(cast.NumOfOps)]
	//for j := i; j < i+int(cast.NumOfOps); j++ {
	//	//transaction = append(transaction, &model.ToOperation(operationsOnWire[j]))
	//}
	//i += int(cast.NumOfOps)
	//	default:
	//		//transaction = []*
	//	}
	//}
}

// CreatePushPullPack ...
func (w *WiredDatatype) CreatePushPullPack() *model.PushPullPack {
	seq := w.checkPoint.Cseq
	operations := w.getOperationOnWires(seq + 1)
	cp := &model.CheckPoint{
		Sseq: w.checkPoint.GetSseq(),
		Cseq: w.checkPoint.GetCseq() + uint64(len(operations)),
	}
	option := model.PushPullBitNormal
	if w.state == model.StateOfDatatype_DUE_TO_CREATE {
		option.SetCreateBit()
	} else if w.state == model.StateOfDatatype_DUE_TO_SUBSCRIBE {
		option.SetSubscribeBit()
	} else if w.state == model.StateOfDatatype_DUE_TO_SUBSCRIBE_CREATE {
		option.SetSubscribeBit().SetCreateBit()
	}
	return &model.PushPullPack{
		Key:        w.Key,
		DUID:       w.id,
		Option:     uint32(option),
		CheckPoint: cp,
		Era:        w.GetEra(),
		Type:       int32(w.TypeOf),
		Operations: operations,
	}
}

func (w *WiredDatatype) calculatePullingOperations(newCheckPoint *model.CheckPoint) int {
	// A: (newCheckPoint.Sseq - w.checkPoint.Sseq) : the number of operations newly pulled, including local pushed operations
	// B: (newCheckPoint.Csseq - w.checkPoint.Cseq) : the number of local operation just pushed
	// A - B: the operations that should be pulled excluding locally pushed operations
	return int((newCheckPoint.Sseq - w.checkPoint.Sseq) - (newCheckPoint.Cseq - w.checkPoint.Cseq))
}

func (w *WiredDatatype) applyPushPullPackExcludeDuplicatedOperations(ppp *model.PushPullPack) {
	pulled := w.calculatePullingOperations(ppp.CheckPoint)
	if len(ppp.Operations) > pulled {
		// for example, if len(ppp.Operations) == 5: o_1 o_2 o_3 o_4 o_5 are received, and
		// if `pulled` == 3, o_1 and o_2 were already received,
		// o_1 and o_2 should be skipped
		skip := len(ppp.Operations) - pulled
		ppp.Operations = ppp.Operations[skip:]
		log.Logger.Infof("skip %s operations", skip)
	}
}

func (w *WiredDatatype) applyPushPullPackSyncCheckPoint(newCheckPoint *model.CheckPoint) {
	oldCheckPoint := w.checkPoint.Clone()
	if w.checkPoint.Cseq < newCheckPoint.Cseq {
		w.checkPoint.Cseq = newCheckPoint.Cseq
	}
	if w.checkPoint.Sseq < newCheckPoint.Sseq {
		w.checkPoint.Sseq = newCheckPoint.Sseq
	}
	log.Logger.Infof("sync CheckPoint: (%+v) -> (%+v)", oldCheckPoint, w.checkPoint)
}

func (w *WiredDatatype) applyPushPullPackUpdateStateOfDatatype(ppp *model.PushPullPack) {
	oldState := w.state
	switch w.state {
	case model.StateOfDatatype_DUE_TO_CREATE,
		model.StateOfDatatype_DUE_TO_SUBSCRIBE,
		model.StateOfDatatype_DUE_TO_SUBSCRIBE_CREATE:
		w.state = model.StateOfDatatype_SUBSCRIBED
		w.id = ppp.DUID
	case model.StateOfDatatype_SUBSCRIBED:
	case model.StateOfDatatype_DUE_TO_UNSUBSCRIBE:
	case model.StateOfDatatype_UNSUBSCRIBED:
	case model.StateOfDatatype_DELETED:
	}
	log.Logger.Infof("update state: %v -> %v", oldState, w.state)
}

// ApplyPushPullPack ...
func (w *WiredDatatype) ApplyPushPullPack(ppp *model.PushPullPack) {
	w.applyPushPullPackExcludeDuplicatedOperations(ppp)
	w.applyPushPullPackSyncCheckPoint(ppp.CheckPoint)
	w.applyPushPullPackUpdateStateOfDatatype(ppp)
	w.applyPushPullPackExecuteOperations(ppp.Operations)
}

func (w *WiredDatatype) getOperationOnWires(cseq uint64) []*model.OperationOnWire {

	if len(w.buffer) == 0 {
		return []*model.OperationOnWire{}
	}
	op := model.ToOperation(w.buffer[0])
	startCseq := op.GetBase().ID.GetSeq()
	var start = int(cseq - startCseq)
	if len(w.buffer) > start {
		return w.buffer[start:]
	}
	return []*model.OperationOnWire{}
}

func (w *WiredDatatype) deliverTransaction(transaction []model.Operation) {
	for _, op := range transaction {
		w.buffer = append(w.buffer, model.ToOperationOnWire(op))
	}
	w.wire.DeliverTransaction(w)
}
