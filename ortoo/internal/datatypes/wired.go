package datatypes

import (
	"fmt"
	"github.com/knowhunger/ortoo/ortoo/errors"
	"github.com/knowhunger/ortoo/ortoo/iface"
	"github.com/knowhunger/ortoo/ortoo/internal/constants"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/ortoo/model"
	"github.com/knowhunger/ortoo/ortoo/operations"
)

// WiredDatatype implements the datatype features related to the synchronization with Ortoo server
type WiredDatatype struct {
	*BaseDatatype
	wire       iface.Wire
	checkPoint *model.CheckPoint
	buffer     []*model.Operation
}

// newWiredDatatype creates a new wiredDatatype
func newWiredDatatype(b *BaseDatatype, w iface.Wire) *WiredDatatype {
	return &WiredDatatype{
		BaseDatatype: b,
		checkPoint:   model.NewCheckPoint(),
		buffer:       make([]*model.Operation, 0, constants.OperationBufferSize),
		wire:         w,
	}
}

// GetBase returns BaseDatatype
func (w *WiredDatatype) GetBase() *BaseDatatype {
	return w.BaseDatatype
}

func (w *WiredDatatype) String() string {
	return w.BaseDatatype.String()
}

// ExecuteRemote ...
func (w *WiredDatatype) ExecuteRemote(op iface.Operation) {
	w.opID.SyncLamport(op.GetID().Lamport)
	w.executeRemoteBase(op)
}

// ReceiveRemoteModelOperations ...
func (w *WiredDatatype) ReceiveRemoteModelOperations(ops []*model.Operation) ([]interface{}, error) {

	datatype := w.datatype
	var opList []interface{}
	for i := 0; i < len(ops); {
		modelOp := ops[i]
		var transaction []*model.Operation
		switch modelOp.GetOpType() {
		case model.TypeOfOperation_TRANSACTION:
			trxOp := operations.ModelToOperation(modelOp).(*operations.TransactionOperation)
			opList = append(opList, trxOp)
			transaction = ops[i : i+int(trxOp.GetNumOfOps())]
			i += int(trxOp.GetNumOfOps())
		default:
			transaction = []*model.Operation{modelOp}
			i++
		}
		trxList, err := datatype.ExecuteTransactionRemote(transaction, true)
		if err != nil {
			return nil, w.Logger.OrtooErrorf(err, "fail to execute Transaction")
		}
		opList = append(opList, trxList)
	}
	return opList, nil
}

func (w *WiredDatatype) applyPushPullPackExecuteOperations(operations []*model.Operation) ([]interface{}, error) {
	return w.ReceiveRemoteModelOperations(operations)
}

// CreatePushPullPack ...
func (w *WiredDatatype) CreatePushPullPack() *model.PushPullPack {
	seq := w.checkPoint.Cseq
	modelOps := w.getModelOperations(seq + 1)
	cp := &model.CheckPoint{
		Sseq: w.checkPoint.GetSseq(),
		Cseq: w.checkPoint.GetCseq() + uint64(len(modelOps)),
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
		Operations: modelOps,
	}
}

func (w *WiredDatatype) calculatePullingOperations(newCheckPoint *model.CheckPoint) int {
	// A: (newCheckPoint.Sseq - w.checkPoint.Sseq) : the number of operations newly pulled, including local pushed operations
	// B: (newCheckPoint.Csseq - w.checkPoint.Cseq) : the number of local operation just pushed
	// A - B: the operations that should be pulled excluding locally pushed operations
	return int((newCheckPoint.Sseq - w.checkPoint.Sseq) - (newCheckPoint.Cseq - w.checkPoint.Cseq))
}

func (w *WiredDatatype) checkPushPullPackOption(ppp *model.PushPullPack) error {
	if ppp.GetPushPullPackOption().HasErrorBit() {
		modelOp := ppp.GetOperations()[0]
		errOp, ok := operations.ModelToOperation(modelOp).(*operations.ErrorOperation)
		if ok {
			switch errOp.GetPushPullError().Code {
			case errors.PushPullErrQueryToDB:
			case errors.PushPullErrIllegalFormat:
			case errors.PushPullErrDuplicateDatatypeKey:
				err := errors.NewDatatypeError(errors.ErrDatatypeCreate, fmt.Sprintf("duplicated key:'%s'", w.Key))
				return err
			case errors.PushPullErrPullOperations:
			case errors.PushPullErrPushOperations:
			case errors.PushPullErrMissingOperations:
			}
		} else {
			panic("Not implemented yet")
		}
	} else if ppp.GetPushPullPackOption().HasSubscribeBit() {

		// modelOp := ppp.GetOperations()[0]
		// snapOp, ok := operations.ModelToOperation(modelOp).(*operations.SnapshotOperation)
		// if ok {
		// 	w.checkPoint = ppp.CheckPoint
		// }
	}
	return nil
}

func (w *WiredDatatype) applyPushPullPackExcludeDuplicatedOperations(ppp *model.PushPullPack) {
	pulled := w.calculatePullingOperations(ppp.CheckPoint)
	if len(ppp.Operations) > pulled {
		// for example, if len(ppp.Operations) == 5: o_1 o_2 o_3 o_4 o_5 are received, and
		// if `pulled` == 3, o_1 and o_2 were already received,
		// o_1 and o_2 should be skipped
		skip := len(ppp.Operations) - pulled
		ppp.Operations = ppp.Operations[skip:]
		log.Logger.Infof("skip %d operations", skip)
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

func (w *WiredDatatype) applyPushPullPackUpdateStateOfDatatype(ppp *model.PushPullPack) (model.StateOfDatatype, model.StateOfDatatype, error) {
	var err error = nil
	oldState := w.state
	switch w.state {
	case model.StateOfDatatype_DUE_TO_CREATE,
		model.StateOfDatatype_DUE_TO_SUBSCRIBE,
		model.StateOfDatatype_DUE_TO_SUBSCRIBE_CREATE:
		if w.state == model.StateOfDatatype_DUE_TO_SUBSCRIBE_CREATE && ppp.GetPushPullPackOption().HasSubscribeBit() {
			w.buffer = make([]*model.Operation, 0, constants.OperationBufferSize)
			newOpID := model.NewOperationIDWithCUID(w.opID.CUID)
			newOpID.Lamport = 1 // Because of SnapshotOperation
			w.SetOpID(newOpID)
			log.Logger.Infof("reset buffer and opID:%s because DUE_TO_SUBSCRIBE_CREATE = > SUBSCRIBE", w.opID.ToString())
		}

		w.state = model.StateOfDatatype_SUBSCRIBED
		w.id = ppp.DUID

		err = w.wire.OnChangeDatatypeState(w.datatype, w.state)
	case model.StateOfDatatype_SUBSCRIBED:
	case model.StateOfDatatype_DUE_TO_UNSUBSCRIBE:
	case model.StateOfDatatype_UNSUBSCRIBED:
	case model.StateOfDatatype_DELETED:
	}
	if oldState != w.state {
		log.Logger.Infof("update state: %v -> %v", oldState, w.state)
	}
	return oldState, w.state, err
}

// ApplyPushPullPack ...
func (w *WiredDatatype) ApplyPushPullPack(ppp *model.PushPullPack) {
	var oldState, newState model.StateOfDatatype
	var errs []error
	var opList []interface{}
	err := w.checkPushPullPackOption(ppp)
	if err == nil {
		w.applyPushPullPackExcludeDuplicatedOperations(ppp)
		w.applyPushPullPackSyncCheckPoint(ppp.CheckPoint)
		oldState, newState, err = w.applyPushPullPackUpdateStateOfDatatype(ppp)
		if err != nil {
			errs = append(errs, err)
		}
		opList, err = w.applyPushPullPackExecuteOperations(ppp.Operations)
		if err != nil {
			errs = append(errs, err)
		}
	} else {
		errs = append(errs, err)
	}
	w.applyPushPullPackCallHandler(errs, oldState, newState, opList)
}

func (w *WiredDatatype) applyPushPullPackCallHandler(errs []error, oldState, newState model.StateOfDatatype, opList []interface{}) {
	if oldState != newState {
		w.datatype.HandleStateChange(oldState, newState)
	}
	if len(errs) > 0 {
		w.datatype.HandleErrors(errs...)
	}
	if len(opList) > 0 {
		w.datatype.HandleRemoteOperations(opList)
	}
}

func (w *WiredDatatype) getModelOperations(cseq uint64) []*model.Operation {

	if len(w.buffer) == 0 {
		return []*model.Operation{}
	}
	op := w.buffer[0]
	startCseq := op.ID.GetSeq()
	var start = int(cseq - startCseq)
	if start >= 0 && len(w.buffer) > start {
		return w.buffer[start:]
	}
	return []*model.Operation{}
}

func (w *WiredDatatype) deliverTransaction(transaction []iface.Operation) {
	if w.wire == nil {
		return
	}
	for _, op := range transaction {
		w.buffer = append(w.buffer, op.ToModelOperation())
	}
	w.wire.DeliverTransaction(w)
}

// NeedSync verifies if the datatype needs to sync
func (w *WiredDatatype) NeedSync(sseq uint64) bool {
	return w.checkPoint.Sseq < sseq
}
