package datatypes

import (
	"fmt"
	"github.com/orda-io/orda/client/pkg/constants"
	errors2 "github.com/orda-io/orda/client/pkg/errors"
	iface2 "github.com/orda-io/orda/client/pkg/iface"
	model2 "github.com/orda-io/orda/client/pkg/model"
	operations2 "github.com/orda-io/orda/client/pkg/operations"
)

// WiredDatatype implements the datatype features related to the synchronization with Orda server
type WiredDatatype struct {
	*TransactionDatatype
	wire        iface2.Wire
	checkPoint  *model2.CheckPoint
	localBuffer []*model2.Operation
}

// NewWiredDatatype creates a new wiredDatatype
func NewWiredDatatype(w iface2.Wire, t *TransactionDatatype) *WiredDatatype {
	return &WiredDatatype{
		TransactionDatatype: t,
		checkPoint:          model2.NewCheckPoint(),
		localBuffer:         make([]*model2.Operation, 0, constants.OperationBufferSize),
		wire:                w,
	}
}

func (its *WiredDatatype) ResetWired() {
	its.localBuffer = make([]*model2.Operation, 0, constants.OperationBufferSize)
	its.opID.Seq = 0
}

func (its *WiredDatatype) String() string {
	return its.BaseDatatype.String()
}

func (its *WiredDatatype) SetCheckPoint(sseq uint64, cseq uint64) {
	its.checkPoint.Sseq = sseq
	its.checkPoint.Cseq = cseq
}

// ReceiveRemoteModelOperations ...
func (its *WiredDatatype) ReceiveRemoteModelOperations(ops []*model2.Operation, obtainList bool) ([]interface{}, errors2.OrdaError) {
	// datatype := its.datatype
	var opList []interface{}
	for i := 0; i < len(ops); {
		modelOp := ops[i]
		var transaction []*model2.Operation
		switch modelOp.GetOpType() {
		case model2.TypeOfOperation_TRANSACTION:
			txOp := operations2.ModelToOperation(modelOp).(*operations2.TransactionOperation)
			opList = append(opList, txOp)
			transaction = ops[i : i+int(txOp.GetNumOfOps())]
			i += int(txOp.GetNumOfOps())
		default:
			transaction = []*model2.Operation{modelOp}
			i++
		}
		txList, err := its.ExecuteRemoteTransaction(transaction, obtainList)
		if err != nil {
			return nil, err
		}
		opList = append(opList, txList)
	}
	return opList, nil
}

// CreatePushPullPack ...
func (its *WiredDatatype) CreatePushPullPack() *model2.PushPullPack {
	seq := its.checkPoint.Cseq
	modelOps := its.getModelOperations(seq + 1)
	cp := &model2.CheckPoint{
		Sseq: its.checkPoint.GetSseq(),
		Cseq: its.checkPoint.GetCseq() + uint64(len(modelOps)),
	}
	option := model2.PushPullBitNormal
	if its.state == model2.StateOfDatatype_DUE_TO_CREATE {
		option.SetCreateBit()
	} else if its.state == model2.StateOfDatatype_DUE_TO_SUBSCRIBE {
		option.SetSubscribeBit()
	} else if its.state == model2.StateOfDatatype_DUE_TO_SUBSCRIBE_CREATE {
		option.SetSubscribeBit().SetCreateBit()
	}
	return &model2.PushPullPack{
		Key:        its.Key,
		DUID:       its.id,
		Option:     uint32(option),
		CheckPoint: cp,
		Era:        its.GetEra(),
		Type:       its.TypeOf,
		Operations: modelOps,
	}
}

func (its *WiredDatatype) getModelOperations(cseq uint64) []*model2.Operation {

	if len(its.localBuffer) == 0 {
		return []*model2.Operation{}
	}
	op := its.localBuffer[0]
	startCseq := op.ID.GetSeq()
	var start = int(cseq - startCseq)
	if start >= 0 && len(its.localBuffer) > start {
		return its.localBuffer[start:]
	}
	// FIXME: return error?
	return []*model2.Operation{}
}

func (its *WiredDatatype) calculatePullingOperations(newCheckPoint *model2.CheckPoint) int {
	// A: (newCheckPoint.Sseq - its.checkPoint.Sseq) : the number of operations newly pulled, including local pushed operations
	// B: (newCheckPoint.Csseq - its.checkPoint.Cseq) : the number of local operation just pushed
	// A - B: the operations that should be pulled excluding locally pushed operations
	return int((newCheckPoint.Sseq - its.checkPoint.Sseq) - (newCheckPoint.Cseq - its.checkPoint.Cseq))
}

func (its *WiredDatatype) checkOptionAndError(ppp *model2.PushPullPack) errors2.OrdaError {
	if ppp.GetPushPullPackOption().HasErrorBit() {
		modelOp := ppp.GetOperations()[0]
		errOp, ok := operations2.ModelToOperation(modelOp).(*operations2.ErrorOperation)
		if ok {
			switch errOp.GetPushPullError().Code {
			case errors2.PushPullAbortionOfServer:
				// TODO: implement me.
			case errors2.PushPullAbortionOfClient:
				// TODO: implement me.
			case errors2.PushPullDuplicateKey:
				return errors2.DatatypeCreate.New(its.L(), fmt.Sprintf("duplicated key:'%s'", its.Key))
			case errors2.PushPullMissingOps:
				// TODO: implement me.
			case errors2.PushPullNoDatatypeToSubscribe:
				return errors2.DatatypeSubscribe.New(its.L(), fmt.Sprintf("%v", errOp.GetPushPullError().Msg))
			}
			panic("Not implemented yet")
		} else {
			panic("Not implemented yet")
		}
	} else if ppp.GetPushPullPackOption().HasSubscribeBit() {
		modelOp := ppp.GetOperations()[0]
		_, ok := operations2.ModelToOperation(modelOp).(*operations2.SnapshotOperation)
		if !ok {
			return errors2.DatatypeSubscribe.New(its.L(), "subscribe without SnapshotOp")
		}
		its.ResetWired()
		its.ResetSnapshot()
		its.ResetTransaction()
		its.checkPoint.Cseq = ppp.CheckPoint.Cseq
		its.checkPoint.Sseq = ppp.CheckPoint.Sseq - uint64(len(ppp.Operations))
		its.L().Infof("ready to subscribe: (%v)", its.checkPoint)
	}
	return nil
}

func (its *WiredDatatype) excludeDuplicatedOperations(ppp *model2.PushPullPack) {
	pulled := its.calculatePullingOperations(ppp.CheckPoint)
	if len(ppp.Operations) > pulled {
		// for example, if len(ppp.Operations) == 5: o_1 o_2 o_3 o_4 o_5 are received, and
		// if `pulled` == 3, o_1 and o_2 were already received,
		// o_1 and o_2 should be skipped
		skip := len(ppp.Operations) - pulled
		ppp.Operations = ppp.Operations[skip:]
		its.L().Infof("skip %d operations", skip)
	}
}

func (its *WiredDatatype) syncCheckPoint(newCheckPoint *model2.CheckPoint) {
	oldCheckPoint := its.checkPoint.Clone()
	if its.checkPoint.Cseq < newCheckPoint.Cseq {
		its.checkPoint.Cseq = newCheckPoint.Cseq
	}
	if its.checkPoint.Sseq < newCheckPoint.Sseq {
		its.checkPoint.Sseq = newCheckPoint.Sseq
	}
	its.L().Infof("sync CheckPoint: (%+v) -> (%+v)", oldCheckPoint, its.checkPoint)
}

func (its *WiredDatatype) updateStateOfDatatype(
	ppp *model2.PushPullPack,
) (model2.StateOfDatatype, model2.StateOfDatatype, errors2.OrdaError) {
	var err errors2.OrdaError = nil
	oldState := its.state
	switch its.state {
	case model2.StateOfDatatype_DUE_TO_CREATE,
		model2.StateOfDatatype_DUE_TO_SUBSCRIBE,
		model2.StateOfDatatype_DUE_TO_SUBSCRIBE_CREATE:
		if its.state == model2.StateOfDatatype_DUE_TO_SUBSCRIBE_CREATE && ppp.GetPushPullPackOption().HasSubscribeBit() {
			its.localBuffer = make([]*model2.Operation, 0, constants.OperationBufferSize)
			newOpID := model2.NewOperationIDWithCUID(its.opID.CUID)
			newOpID.Lamport = 1 // Because of SnapshotOperation
			its.SetOpID(newOpID)
			its.L().Infof("reset buffer and opID:%s because DUE_TO_SUBSCRIBE_CREATE = > SUBSCRIBE", its.opID.ToString())
		}

		its.state = model2.StateOfDatatype_SUBSCRIBED
		its.id = ppp.DUID

		err = its.wire.OnChangeDatatypeState(its.Datatype, its.state)
	case model2.StateOfDatatype_SUBSCRIBED:
	case model2.StateOfDatatype_DUE_TO_UNSUBSCRIBE:
	case model2.StateOfDatatype_CLOSED:
	case model2.StateOfDatatype_DELETED:
	}
	if oldState != its.state {
		its.L().Infof("update state: %v -> %v", oldState, its.state)
	}
	return oldState, its.state, err
}

// ApplyPushPullPack ...
func (its *WiredDatatype) ApplyPushPullPack(ppp *model2.PushPullPack) {
	its.L().Infof("begin ApplyPushPull:%v", ppp.ToString())
	defer its.L().Infof("end ApplyPushPull")
	var oldState, newState model2.StateOfDatatype
	var errs errors2.OrdaError = &errors2.MultipleOrdaErrors{}
	var opList []interface{}
	err := its.checkOptionAndError(ppp)
	if err == nil {
		its.excludeDuplicatedOperations(ppp)
		its.syncCheckPoint(ppp.CheckPoint)
		oldState, newState, err = its.updateStateOfDatatype(ppp)
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
	go its.callHandlers(errs, oldState, newState, opList)
}

func (its *WiredDatatype) callHandlers(
	errs errors2.OrdaError,
	oldState,
	newState model2.StateOfDatatype,
	opList []interface{},
) {
	if oldState != newState {
		its.HandleStateChange(oldState, newState)
	}
	if errs.Size() > 0 {
		its.HandleErrors(errs.ToArray()...)
	}
	if len(opList) > 0 {
		its.HandleRemoteOperations(opList)
	}
}

func (its *WiredDatatype) DeliverTransaction(transaction []iface2.Operation) {

	for _, op := range transaction {
		its.localBuffer = append(its.localBuffer, op.ToModelOperation())
	}
	if its.wire == nil && its.ctx.Client.SyncType != model2.SyncType_REALTIME {
		return
	}
	its.ctx.L().Infof("call deliverTransaction: %v", transaction)
	its.wire.DeliverTransaction(its)
}

// NeedPull verifies if the datatype needs to pull
func (its *WiredDatatype) NeedPull(sseq uint64) bool {
	return its.checkPoint.Sseq < sseq
}

// NeedPush verifies if the datatype needs to push
func (its *WiredDatatype) NeedPush() bool {
	return its.checkPoint.Cseq < its.opID.GetSeq()
}
