package datatypes

import (
	"fmt"
	"github.com/knowhunger/ortoo/pkg/constants"
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/iface"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/pkg/operations"
)

// WiredDatatype implements the datatype features related to the synchronization with Ortoo server
type WiredDatatype struct {
	*BaseDatatype
	wire        iface.Wire
	checkPoint  *model.CheckPoint
	localBuffer []*model.Operation
}

// newWiredDatatype creates a new wiredDatatype
func newWiredDatatype(b *BaseDatatype, w iface.Wire) *WiredDatatype {
	return &WiredDatatype{
		BaseDatatype: b,
		checkPoint:   model.NewCheckPoint(),
		localBuffer:  make([]*model.Operation, 0, constants.OperationBufferSize),
		wire:         w,
	}
}

func (its *WiredDatatype) ResetWired() {
	its.localBuffer = make([]*model.Operation, 0, constants.OperationBufferSize)
}

// GetBase returns BaseDatatype
func (its *WiredDatatype) GetBase() *BaseDatatype {
	return its.BaseDatatype
}

func (its *WiredDatatype) String() string {
	return its.BaseDatatype.String()
}

// ExecuteRemote ...
func (its *WiredDatatype) ExecuteRemote(op iface.Operation) {
	its.opID.SyncLamport(op.GetID().Lamport) // TODO: this should move to baseDatatype
	its.executeRemoteBase(op)
}

// ReceiveRemoteModelOperations ...
func (its *WiredDatatype) ReceiveRemoteModelOperations(ops []*model.Operation, obtainList bool) ([]interface{}, errors.OrtooError) {
	datatype := its.datatype
	var opList []interface{}
	for i := 0; i < len(ops); {
		modelOp := ops[i]
		var transaction []*model.Operation
		switch modelOp.GetOpType() {
		case model.TypeOfOperation_TRANSACTION:
			txOp := operations.ModelToOperation(modelOp).(*operations.TransactionOperation)
			opList = append(opList, txOp)
			transaction = ops[i : i+int(txOp.GetNumOfOps())]
			i += int(txOp.GetNumOfOps())
		default:
			transaction = []*model.Operation{modelOp}
			i++
		}
		txList, err := datatype.ExecuteRemoteTransaction(transaction, obtainList)
		if err != nil {
			return nil, err
		}
		opList = append(opList, txList)
	}
	return opList, nil
}

// CreatePushPullPack ...
func (its *WiredDatatype) CreatePushPullPack() *model.PushPullPack {
	seq := its.checkPoint.Cseq
	modelOps := its.getModelOperations(seq + 1)
	cp := &model.CheckPoint{
		Sseq: its.checkPoint.GetSseq(),
		Cseq: its.checkPoint.GetCseq() + uint64(len(modelOps)),
	}
	option := model.PushPullBitNormal
	if its.state == model.StateOfDatatype_DUE_TO_CREATE {
		option.SetCreateBit()
	} else if its.state == model.StateOfDatatype_DUE_TO_SUBSCRIBE {
		option.SetSubscribeBit()
	} else if its.state == model.StateOfDatatype_DUE_TO_SUBSCRIBE_CREATE {
		option.SetSubscribeBit().SetCreateBit()
	}
	return &model.PushPullPack{
		Key:        its.Key,
		DUID:       its.id,
		Option:     uint32(option),
		CheckPoint: cp,
		Era:        its.GetEra(),
		Type:       its.TypeOf,
		Operations: modelOps,
	}
}

func (its *WiredDatatype) getModelOperations(cseq uint64) []*model.Operation {

	if len(its.localBuffer) == 0 {
		return []*model.Operation{}
	}
	op := its.localBuffer[0]
	startCseq := op.ID.GetSeq()
	var start = int(cseq - startCseq)
	if start >= 0 && len(its.localBuffer) > start {
		return its.localBuffer[start:]
	}
	// FIXME: return error?
	return []*model.Operation{}
}

func (its *WiredDatatype) calculatePullingOperations(newCheckPoint *model.CheckPoint) int {
	// A: (newCheckPoint.Sseq - its.checkPoint.Sseq) : the number of operations newly pulled, including local pushed operations
	// B: (newCheckPoint.Csseq - its.checkPoint.Cseq) : the number of local operation just pushed
	// A - B: the operations that should be pulled excluding locally pushed operations
	return int((newCheckPoint.Sseq - its.checkPoint.Sseq) - (newCheckPoint.Cseq - its.checkPoint.Cseq))
}

func (its *WiredDatatype) checkPushPullPackOption(ppp *model.PushPullPack) errors.OrtooError {
	if ppp.GetPushPullPackOption().HasErrorBit() {
		modelOp := ppp.GetOperations()[0]
		errOp, ok := operations.ModelToOperation(modelOp).(*operations.ErrorOperation)
		if ok {
			switch errOp.GetPushPullError().Code {
			case errors.PushPullAbortedForServer:
				// TODO: implement me.
			case errors.PushPullAbortedForClient:
				// TODO: implement me.
			case errors.PushPullDuplicateKey:
				return errors.DatatypeCreate.New(its.Logger, fmt.Sprintf("duplicated key:'%s'", its.Key))
			case errors.PushPullMissingOps:
				// TODO: implement me.
			}
		} else {
			panic("Not implemented yet")
		}
	} else if ppp.GetPushPullPackOption().HasSubscribeBit() {

		modelOp := ppp.GetOperations()[0]
		_, ok := operations.ModelToOperation(modelOp).(*operations.SnapshotOperation)
		if !ok {
			return errors.DatatypeSubscribe.New(its.Logger, "subscribe without SnapshotOp")
		}
		its.datatype.ResetRollBackContext()
		its.checkPoint.Cseq = ppp.CheckPoint.Cseq
		its.checkPoint.Sseq = ppp.CheckPoint.Sseq - uint64(len(ppp.Operations))
		its.Logger.Infof("ready to subscribe: (%v)", its.checkPoint)
	}
	return nil
}

func (its *WiredDatatype) applyPushPullPackExcludeDuplicatedOperations(ppp *model.PushPullPack) {
	pulled := its.calculatePullingOperations(ppp.CheckPoint)
	if len(ppp.Operations) > pulled {
		// for example, if len(ppp.Operations) == 5: o_1 o_2 o_3 o_4 o_5 are received, and
		// if `pulled` == 3, o_1 and o_2 were already received,
		// o_1 and o_2 should be skipped
		skip := len(ppp.Operations) - pulled
		ppp.Operations = ppp.Operations[skip:]
		its.Logger.Infof("skip %d operations", skip)
	}
}

func (its *WiredDatatype) applyPushPullPackSyncCheckPoint(newCheckPoint *model.CheckPoint) {
	oldCheckPoint := its.checkPoint.Clone()
	if its.checkPoint.Cseq < newCheckPoint.Cseq {
		its.checkPoint.Cseq = newCheckPoint.Cseq
	}
	if its.checkPoint.Sseq < newCheckPoint.Sseq {
		its.checkPoint.Sseq = newCheckPoint.Sseq
	}
	its.Logger.Infof("sync CheckPoint: (%+v) -> (%+v)", oldCheckPoint, its.checkPoint)
}

func (its *WiredDatatype) applyPushPullPackUpdateStateOfDatatype(
	ppp *model.PushPullPack,
) (model.StateOfDatatype, model.StateOfDatatype, errors.OrtooError) {
	var err errors.OrtooError = nil
	oldState := its.state
	switch its.state {
	case model.StateOfDatatype_DUE_TO_CREATE,
		model.StateOfDatatype_DUE_TO_SUBSCRIBE,
		model.StateOfDatatype_DUE_TO_SUBSCRIBE_CREATE:
		if its.state == model.StateOfDatatype_DUE_TO_SUBSCRIBE_CREATE && ppp.GetPushPullPackOption().HasSubscribeBit() {
			its.localBuffer = make([]*model.Operation, 0, constants.OperationBufferSize)
			newOpID := model.NewOperationIDWithCUID(its.opID.CUID)
			newOpID.Lamport = 1 // Because of SnapshotOperation
			its.SetOpID(newOpID)
			its.Logger.Infof("reset buffer and opID:%s because DUE_TO_SUBSCRIBE_CREATE = > SUBSCRIBE", its.opID.ToString())
		}

		its.state = model.StateOfDatatype_SUBSCRIBED
		its.id = ppp.DUID

		err = its.wire.OnChangeDatatypeState(its.datatype, its.state)
	case model.StateOfDatatype_SUBSCRIBED:
	case model.StateOfDatatype_DUE_TO_UNSUBSCRIBE:
	case model.StateOfDatatype_UNSUBSCRIBED:
	case model.StateOfDatatype_DELETED:
	}
	if oldState != its.state {
		its.Logger.Infof("update state: %v -> %v", oldState, its.state)
	}
	return oldState, its.state, err
}

// ApplyPushPullPack ...
func (its *WiredDatatype) ApplyPushPullPack(ppp *model.PushPullPack) {
	its.Logger.Infof("begin ApplyPushPull:%v", ppp.ToString())
	defer its.Logger.Infof("end ApplyPushPull")
	var oldState, newState model.StateOfDatatype
	var errs errors.OrtooError = &errors.MultipleOrtooErrors{}
	var opList []interface{}
	err := its.checkPushPullPackOption(ppp)
	if err == nil {
		its.applyPushPullPackExcludeDuplicatedOperations(ppp)
		its.applyPushPullPackSyncCheckPoint(ppp.CheckPoint)
		oldState, newState, err = its.applyPushPullPackUpdateStateOfDatatype(ppp)
		if err != nil {
			errs = errs.Append(err)
		}
		opList, err = its.ReceiveRemoteModelOperations(ppp.Operations, true)
		if err != nil {
			errs = errs.Append(err)
		}
	} else {
		errs = errs.Append(err)
	}
	go its.applyPushPullPackCallHandler(errs, oldState, newState, opList)
}

func (its *WiredDatatype) applyPushPullPackCallHandler(
	errs errors.OrtooError,
	oldState,
	newState model.StateOfDatatype,
	opList []interface{},
) {
	if oldState != newState {
		its.datatype.HandleStateChange(oldState, newState)
	}
	if errs.Size() > 0 {
		its.datatype.HandleErrors(errs.ToArray()...)
	}
	if len(opList) > 0 {
		its.datatype.HandleRemoteOperations(opList)
	}
}

func (its *WiredDatatype) deliverTransaction(transaction []iface.Operation) {
	if its.wire == nil {
		return
	}
	for _, op := range transaction {
		its.localBuffer = append(its.localBuffer, op.ToModelOperation())
	}
	its.wire.DeliverTransaction(its)
}

// NeedSync verifies if the datatype needs to sync
func (its *WiredDatatype) NeedSync(sseq uint64) bool {
	return its.checkPoint.Sseq < sseq
}
