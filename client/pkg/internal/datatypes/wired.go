package datatypes

import (
	"fmt"
	"github.com/orda-io/orda/client/pkg/constants"
	"github.com/orda-io/orda/client/pkg/errors"
	"github.com/orda-io/orda/client/pkg/iface"
	"github.com/orda-io/orda/client/pkg/model"
	"github.com/orda-io/orda/client/pkg/operations"
)

// WiredDatatype implements the datatype features related to the synchronization with Orda server
type WiredDatatype struct {
	*TransactionDatatype
	wire        iface.Wire
	checkPoint  *model.CheckPoint
	localBuffer []*model.Operation
}

// NewWiredDatatype creates a new wiredDatatype
func NewWiredDatatype(w iface.Wire, t *TransactionDatatype) *WiredDatatype {
	return &WiredDatatype{
		TransactionDatatype: t,
		checkPoint:          model.NewCheckPoint(),
		localBuffer:         make([]*model.Operation, 0, constants.OperationBufferSize),
		wire:                w,
	}
}

// ResetWired resets the data related to WiredDatatype
func (its *WiredDatatype) ResetWired() {
	its.localBuffer = make([]*model.Operation, 0, constants.OperationBufferSize)
	its.opID.Seq = 0
}

// SetCheckPoint sets the CheckPoint
func (its *WiredDatatype) SetCheckPoint(sseq uint64, cseq uint64) {
	its.checkPoint.Sseq = sseq
	its.checkPoint.Cseq = cseq
}

// ReceiveRemoteModelOperations executes remote model operations.
func (its *WiredDatatype) ReceiveRemoteModelOperations(ops []*model.Operation, obtainList bool) ([]interface{}, errors.OrdaError) {
	// datatype := its.datatype
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
		txList, err := its.ExecuteRemoteTransaction(transaction, obtainList)
		if err != nil {
			return nil, err
		}
		opList = append(opList, txList)
	}
	return opList, nil
}

// CreatePushPullPack creates a PushPullPack
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

func (its *WiredDatatype) checkOptionAndError(ppp *model.PushPullPack) errors.OrdaError {
	if ppp.GetPushPullPackOption().HasErrorBit() {
		modelOp := ppp.GetOperations()[0]
		errOp, ok := operations.ModelToOperation(modelOp).(*operations.ErrorOperation)
		if ok {
			switch errOp.GetPushPullError().Code {
			case errors.PushPullAbortionOfServer:
				// TODO: implement me.
			case errors.PushPullAbortionOfClient:
				// TODO: implement me.
			case errors.PushPullDuplicateKey:
				return errors.DatatypeCreate.New(its.L(), fmt.Sprintf("duplicated key:'%s'", its.Key))
			case errors.PushPullMissingOps:
				// TODO: implement me.
			case errors.PushPullNoDatatypeToSubscribe:
				return errors.DatatypeSubscribe.New(its.L(), fmt.Sprintf("%v", errOp.GetPushPullError().Msg))
			}
			panic("Not implemented yet")
		} else {
			panic("Not implemented yet")
		}
	} else if ppp.GetPushPullPackOption().HasSubscribeBit() {
		modelOp := ppp.GetOperations()[0]
		_, ok := operations.ModelToOperation(modelOp).(*operations.SnapshotOperation)
		if !ok {
			return errors.DatatypeSubscribe.New(its.L(), "subscribe without SnapshotOp")
		}
		its.ResetWired()
		its.ResetSnapshot()
		its.ResetTransaction()
		its.checkPoint.Cseq = ppp.CheckPoint.Cseq
		its.checkPoint.Sseq = ppp.CheckPoint.Sseq - uint64(len(ppp.Operations))
		its.L().Infof("ready to subscribe: %s", its.checkPoint.ToString())
	}
	return nil
}

func (its *WiredDatatype) excludeDuplicatedOperations(ppp *model.PushPullPack) {
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

func (its *WiredDatatype) syncCheckPoint(newCheckPoint *model.CheckPoint) {
	oldCheckPoint := its.checkPoint.Clone()
	if its.checkPoint.Cseq < newCheckPoint.Cseq {
		its.checkPoint.Cseq = newCheckPoint.Cseq
	}
	if its.checkPoint.Sseq < newCheckPoint.Sseq {
		its.checkPoint.Sseq = newCheckPoint.Sseq
	}
	its.L().Infof("sync CheckPoint: %s -> %s", oldCheckPoint.ToString(), its.checkPoint.ToString())
}

func (its *WiredDatatype) updateStateOfDatatype(
	ppp *model.PushPullPack,
) (model.StateOfDatatype, model.StateOfDatatype, errors.OrdaError) {
	var err errors.OrdaError = nil
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
			its.L().Infof("reset buffer and opID:%s because DUE_TO_SUBSCRIBE_CREATE = > SUBSCRIBE", its.opID.ToString())
		}

		its.state = model.StateOfDatatype_SUBSCRIBED
		its.id = ppp.DUID

		err = its.wire.OnChangeDatatypeState(its.Datatype, its.state)
	case model.StateOfDatatype_SUBSCRIBED:
	case model.StateOfDatatype_DUE_TO_UNSUBSCRIBE:
	case model.StateOfDatatype_CLOSED:
	case model.StateOfDatatype_DELETED:
	}
	if oldState != its.state {
		its.L().Infof("update state: %v -> %v", oldState, its.state)
	}
	return oldState, its.state, err
}

// ApplyPushPullPack applies for PushPullPack
func (its *WiredDatatype) ApplyPushPullPack(ppp *model.PushPullPack) {
	defer its.L().Infof("end ApplyPushPull")
	var oldState, newState model.StateOfDatatype
	var errs errors.OrdaError = &errors.MultipleOrdaErrors{}
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
	errs errors.OrdaError,
	oldState,
	newState model.StateOfDatatype,
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

// DeliverTransaction delivers the transaction if needed
func (its *WiredDatatype) DeliverTransaction(transaction []iface.Operation) {

	for _, op := range transaction {
		its.localBuffer = append(its.localBuffer, op.ToModelOperation())
	}
	if its.wire == nil && its.ctx.Client.SyncType != model.SyncType_REALTIME {
		return
	}
	its.ctx.L().Debugf("call deliverTransaction: %v", transaction)
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
