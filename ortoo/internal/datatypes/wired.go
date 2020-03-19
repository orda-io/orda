package datatypes

import (
	"fmt"
	"github.com/knowhunger/ortoo/ortoo/errors"
	"github.com/knowhunger/ortoo/ortoo/internal/constants"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/ortoo/model"
)

// Wire defines the interfaces related to delivering operations. This is called when a datatype needs to send messages
type Wire interface {
	DeliverTransaction(wired *WiredDatatype)
	OnChangeDatatypeState(dt model.Datatype, state model.StateOfDatatype) error
}

// WiredDatatype implements the datatype features related to the synchronization with Ortoo server
type WiredDatatype struct {
	*BaseDatatype
	wire       Wire
	checkPoint *model.CheckPoint
	buffer     []*model.OperationOnWire
}

// PublicWiredDatatypeInterface defines the interface related to the synchronization with Ortoo server
type PublicWiredDatatypeInterface interface {
	PublicBaseDatatypeInterface
}

// WiredDatatypeInterface defines the internal interface related to the synchronization with Ortoo server
type WiredDatatypeInterface interface {
	GetBase() *BaseDatatype
	ExecuteRemote(op model.Operation)
	ReceiveRemoteOperationsOnWire(operations []*model.OperationOnWire) error
	ApplyPushPullPack(*model.PushPullPack)
	CreatePushPullPack() *model.PushPullPack
}

// newWiredDatatype creates a new wiredDatatype
func newWiredDatatype(b *BaseDatatype, w Wire) *WiredDatatype {
	return &WiredDatatype{
		BaseDatatype: b,
		checkPoint:   model.NewCheckPoint(),
		buffer:       make([]*model.OperationOnWire, 0, constants.OperationBufferSize),
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
func (w *WiredDatatype) ExecuteRemote(op model.Operation) {
	w.opID.SyncLamport(op.GetBase().GetID().Lamport)
	w.executeRemoteBase(op)
}

// ReceiveRemoteOperationsOnWire ...
func (w *WiredDatatype) ReceiveRemoteOperationsOnWire(operations []*model.OperationOnWire) ([]interface{}, error) {

	finalDatatype := w.datatype
	var opList []interface{}
	for i := 0; i < len(operations); {
		op := model.ToOperation(operations[i])
		opList = append(opList, op.GetAsJSON())
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
			return nil, w.Logger.OrtooErrorf(err, "fail to execute Transaction")
		}
	}
	return opList, nil
}

func (w *WiredDatatype) applyPushPullPackExecuteOperations(operationsOnWire []*model.OperationOnWire) ([]interface{}, error) {
	return w.ReceiveRemoteOperationsOnWire(operationsOnWire)
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

func (w *WiredDatatype) checkPushPullPackOption(ppp *model.PushPullPack) error {
	if ppp.GetPushPullPackOption().HasErrorBit() {
		opOnWire := ppp.GetOperations()[0]
		errOp, ok := model.ToOperation(opOnWire).(*model.ErrorOperation)
		if ok {
			switch errOp.GetPushPullError().Code {
			case model.PushPullErrQueryToDB:
			case model.PushPullErrIllegalFormat:
			case model.PushPullErrDuplicateDatatypeKey:
				err := errors.NewDatatypeError(errors.ErrDatatypeCreate, fmt.Sprintf("duplicated key:'%s'", w.Key))
				return err
			case model.PushPullErrPullOperations:
			case model.PushPullErrPushOperations:
			case model.PushPullErrMissingOperations:
			}
		} else {

		}

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
	if w.wire == nil {
		return
	}
	for _, op := range transaction {
		w.buffer = append(w.buffer, model.ToOperationOnWire(op))
	}
	w.wire.DeliverTransaction(w)
}

// NeedSync verifies if the datatype needs to sync
func (w *WiredDatatype) NeedSync(sseq uint64) bool {
	return w.checkPoint.Sseq < sseq
}
