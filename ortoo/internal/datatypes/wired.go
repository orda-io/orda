package datatypes

import (
	"fmt"
	"github.com/knowhunger/ortoo/ortoo/errors"
	"github.com/knowhunger/ortoo/ortoo/iface"
	"github.com/knowhunger/ortoo/ortoo/internal/constants"
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
func (its *WiredDatatype) GetBase() *BaseDatatype {
	return its.BaseDatatype
}

func (its *WiredDatatype) String() string {
	return its.BaseDatatype.String()
}

// ExecuteRemote ...
func (its *WiredDatatype) ExecuteRemote(op iface.Operation) {
	its.opID.SyncLamport(op.GetID().Lamport)
	its.executeRemoteBase(op)
}

// ReceiveRemoteModelOperations ...
func (its *WiredDatatype) ReceiveRemoteModelOperations(ops []*model.Operation) ([]interface{}, error) {

	datatype := its.datatype
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
			return nil, its.Logger.OrtooErrorf(err, "fail to execute Transaction")
		}
		opList = append(opList, trxList)
	}
	return opList, nil
}

func (its *WiredDatatype) applyPushPullPackExecuteOperations(operations []*model.Operation) ([]interface{}, error) {
	return its.ReceiveRemoteModelOperations(operations)
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
		Type:       int32(its.TypeOf),
		Operations: modelOps,
	}
}

func (its *WiredDatatype) calculatePullingOperations(newCheckPoint *model.CheckPoint) int {
	// A: (newCheckPoint.Sseq - its.checkPoint.Sseq) : the number of operations newly pulled, including local pushed operations
	// B: (newCheckPoint.Csseq - its.checkPoint.Cseq) : the number of local operation just pushed
	// A - B: the operations that should be pulled excluding locally pushed operations
	return int((newCheckPoint.Sseq - its.checkPoint.Sseq) - (newCheckPoint.Cseq - its.checkPoint.Cseq))
}

func (its *WiredDatatype) checkPushPullPackOption(ppp *model.PushPullPack) error {
	if ppp.GetPushPullPackOption().HasErrorBit() {
		modelOp := ppp.GetOperations()[0]
		errOp, ok := operations.ModelToOperation(modelOp).(*operations.ErrorOperation)
		if ok {
			switch errOp.GetPushPullError().Code {
			case errors.PushPullErrQueryToDB:
			case errors.PushPullErrIllegalFormat:
			case errors.PushPullErrDuplicateDatatypeKey:
				err := errors.ErrDatatypeCreate.New(fmt.Sprintf("duplicated key:'%s'", its.Key))
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
		// 	its.checkPoint = ppp.CheckPoint
		// }
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

func (its *WiredDatatype) applyPushPullPackUpdateStateOfDatatype(ppp *model.PushPullPack) (model.StateOfDatatype, model.StateOfDatatype, error) {
	var err error = nil
	oldState := its.state
	switch its.state {
	case model.StateOfDatatype_DUE_TO_CREATE,
		model.StateOfDatatype_DUE_TO_SUBSCRIBE,
		model.StateOfDatatype_DUE_TO_SUBSCRIBE_CREATE:
		if its.state == model.StateOfDatatype_DUE_TO_SUBSCRIBE_CREATE && ppp.GetPushPullPackOption().HasSubscribeBit() {
			its.buffer = make([]*model.Operation, 0, constants.OperationBufferSize)
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
	var oldState, newState model.StateOfDatatype
	var errs []error
	var opList []interface{}
	err := its.checkPushPullPackOption(ppp)
	if err == nil {
		its.applyPushPullPackExcludeDuplicatedOperations(ppp)
		its.applyPushPullPackSyncCheckPoint(ppp.CheckPoint)
		oldState, newState, err = its.applyPushPullPackUpdateStateOfDatatype(ppp)
		if err != nil {
			errs = append(errs, err)
		}
		opList, err = its.applyPushPullPackExecuteOperations(ppp.Operations)
		if err != nil {
			errs = append(errs, err)
		}
	} else {
		errs = append(errs, err)
	}
	its.applyPushPullPackCallHandler(errs, oldState, newState, opList)
}

func (its *WiredDatatype) applyPushPullPackCallHandler(errs []error, oldState, newState model.StateOfDatatype, opList []interface{}) {
	if oldState != newState {
		its.datatype.HandleStateChange(oldState, newState)
	}
	if len(errs) > 0 {
		its.datatype.HandleErrors(errs...)
	}
	if len(opList) > 0 {
		its.datatype.HandleRemoteOperations(opList)
	}
}

func (its *WiredDatatype) getModelOperations(cseq uint64) []*model.Operation {

	if len(its.buffer) == 0 {
		return []*model.Operation{}
	}
	op := its.buffer[0]
	startCseq := op.ID.GetSeq()
	var start = int(cseq - startCseq)
	if start >= 0 && len(its.buffer) > start {
		return its.buffer[start:]
	}
	return []*model.Operation{}
}

func (its *WiredDatatype) deliverTransaction(transaction []iface.Operation) {
	if its.wire == nil {
		return
	}
	for _, op := range transaction {
		its.buffer = append(its.buffer, op.ToModelOperation())
	}
	its.wire.DeliverTransaction(its)
}

// NeedSync verifies if the datatype needs to sync
func (its *WiredDatatype) NeedSync(sseq uint64) bool {
	return its.checkPoint.Sseq < sseq
}
