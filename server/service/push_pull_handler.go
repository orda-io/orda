package service

import (
	"context"
	"fmt"
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/log"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/pkg/operations"
	"github.com/knowhunger/ortoo/pkg/types"
	"github.com/knowhunger/ortoo/server/constants"
	"github.com/knowhunger/ortoo/server/mongodb"
	"github.com/knowhunger/ortoo/server/mongodb/schema"
	"github.com/knowhunger/ortoo/server/notification"
	"github.com/knowhunger/ortoo/server/snapshot"
	"time"
)

type pushPullCase uint32

const (
	caseError pushPullCase = iota
	caseMatchNothing
	caseUsedDUID
	caseMatchKeyNotType
	caseAllMatchedSubscribed
	caseAllMatchedNotSubscribed
	caseAllMatchedNotVisible
)

var (
	pushPullCaseMap = map[pushPullCase]string{
		caseError:                   "caseError",
		caseMatchNothing:            "caseMatchNothing",
		caseUsedDUID:                "caseUsedDUID",
		caseMatchKeyNotType:         "caseMatchKeyNotType",
		caseAllMatchedSubscribed:    "caseAllMatchedSubscribed",
		caseAllMatchedNotSubscribed: "caseAllMatchedNotSubscribed",
		caseAllMatchedNotVisible:    "caseAllMatchedNotVisible",
	}
)

// PushPullHandler is a struct that handles a push-pull for a datatype
type PushPullHandler struct {
	Key  string
	DUID string
	CUID string

	err      *errors.PushPullError
	ctx      context.Context
	mongo    *mongodb.RepositoryMongo
	notifier *notification.Notifier

	initialCheckPoint *model.CheckPoint
	currentCheckPoint *model.CheckPoint

	clientDoc     *schema.ClientDoc
	datatypeDoc   *schema.DatatypeDoc
	collectionDoc *schema.CollectionDoc

	gotPushPullPack *model.PushPullPack
	gotOption       *model.PushPullPackOption

	responsePushPullPack *model.PushPullPack
	retCh                chan *model.PushPullPack

	pushingOperations []interface{}
	pulledOperations  []model.Operation
}

// Start begins the push pull for a datatype and returns the result with the channel 'retCh'
func (its *PushPullHandler) Start() <-chan *model.PushPullPack {
	retCh := make(chan *model.PushPullPack)
	go its.process(retCh)
	return retCh
}

func (its *PushPullHandler) getPushPullTag() errors.PushPullTag {
	return errors.PushPullTag{
		CollectionName: its.collectionDoc.Name,
		Key:            its.Key,
		DUID:           its.DUID,
	}
}

func (its *PushPullHandler) initialize(retCh chan *model.PushPullPack) *errors.PushPullError {
	its.retCh = retCh
	its.responsePushPullPack = its.gotPushPullPack.GetResponsePushPullPack()
	its.responsePushPullPack.Option = uint32(model.PushPullBitNormal)

	checkPoint, err := its.mongo.GetCheckPointFromClient(its.ctx, its.CUID, its.DUID)
	if err != nil {
		return errors.NewPushPullError(errors.PushPullErrQueryToDB, its.getPushPullTag(), err)
	}
	if checkPoint == nil {
		checkPoint = model.NewCheckPoint()
	}

	its.initialCheckPoint = checkPoint.Clone()
	its.currentCheckPoint = checkPoint.Clone()
	return nil
}

func (its *PushPullHandler) finalize() {
	if its.err == nil {
		log.Logger.Infof("finish PUSHPULL for %s (%v) -> (%v)",
			its.datatypeDoc, its.initialCheckPoint, its.currentCheckPoint)
		if len(its.pushingOperations) > 0 {
			if err := its.sendNotification(); err == nil {
				// continue
			}
			if err := its.reserveUpdateSnapshot(); err != nil {
				// continue
			}
		}
	} else {
		its.responsePushPullPack.GetPushPullPackOption().SetErrorBit()

		errOp := operations.NewErrorOperation(its.err)
		its.responsePushPullPack.Operations = append(its.responsePushPullPack.Operations, errOp.ToModelOperation())

	}
	log.Logger.Infof("send back to %s %s", its.clientDoc.Alias, its.responsePushPullPack.ToString())
	its.retCh <- its.responsePushPullPack
}

func (its *PushPullHandler) logInitialConditions(casePushPull string) {
	log.Logger.Infof("initial condition| case: %s, opt: %s, cp(%v), ops:%d, sseqEnd:%d",
		casePushPull, its.gotOption.String(), its.initialCheckPoint, len(its.gotPushPullPack.Operations),
		its.datatypeDoc.SseqEnd,
	)
}

func (its *PushPullHandler) process(retCh chan *model.PushPullPack) {
	defer its.finalize()

	if its.err = its.initialize(retCh); its.err != nil {
		return
	}

	casePushPull, err := its.evaluatePushPullCase()
	if its.err = err; err != nil {
		return
	}

	if its.err = its.processSubscribeOrCreate(casePushPull); its.err != nil {
		return
	}

	its.logInitialConditions(pushPullCaseMap[casePushPull])

	if its.err = its.pushOperations(); its.err != nil {
		return
	}
	if its.err = its.pullOperations(); its.err != nil {
		return
	}
	if its.err = its.commitToMongoDB(); its.err != nil {
		return
	}
	return
}

func (its *PushPullHandler) sendNotification() error {
	if err := its.notifier.NotifyAfterPushPull(
		its.collectionDoc.Name,
		its.clientDoc,
		its.datatypeDoc,
		its.currentCheckPoint.Sseq); err != nil {
		return log.OrtooError(err)
	}

	return nil
}

func (its *PushPullHandler) reserveUpdateSnapshot() error {
	snapshotManager := snapshot.NewManager(its.ctx, its.mongo, its.datatypeDoc, its.collectionDoc)
	if err := snapshotManager.UpdateSnapshot(); err != nil { // TODO: should be asynchronous
		return log.OrtooError(err)
	}
	return nil
}

func (its *PushPullHandler) commitToMongoDB() *errors.PushPullError {
	its.datatypeDoc.SseqEnd = its.currentCheckPoint.Sseq
	its.responsePushPullPack.CheckPoint = its.currentCheckPoint

	if len(its.pushingOperations) > 0 {
		if err := its.mongo.InsertOperations(its.ctx, its.pushingOperations); err != nil {
			return errors.NewPushPullError(errors.PushPullErrQueryToDB, its.getPushPullTag(), err)
		}
		log.Logger.Infof("[MONGO] push %d operations finally", len(its.pushingOperations))
	}

	if err := its.mongo.UpdateDatatype(its.ctx, its.datatypeDoc); err != nil {
		return errors.NewPushPullError(errors.PushPullErrQueryToDB, its.getPushPullTag(), err)
	}
	log.Logger.Infof("[MONGO] update %s", its.datatypeDoc)

	if err := its.mongo.UpdateCheckPointInClient(its.ctx, its.CUID, its.DUID, its.currentCheckPoint); err != nil {
		return errors.NewPushPullError(errors.PushPullErrQueryToDB, its.getPushPullTag(), err)
	}
	log.Logger.Infof("[MONGO] update %s with CP(%s)", its.clientDoc, its.currentCheckPoint.String())
	return nil
}

func (its *PushPullHandler) pullOperations() *errors.PushPullError {
	sseqBegin := its.gotPushPullPack.CheckPoint.Sseq + 1

	var operations []*model.Operation
	if its.datatypeDoc.SseqBegin <= sseqBegin && !its.gotOption.HasSnapshotBit() {
		if err := its.mongo.GetOperations(its.ctx,
			its.DUID,
			sseqBegin,
			constants.InfinitySseq,
			func(opDoc *schema.OperationDoc) error {
				var modelOp = opDoc.GetOperation()
				operations = append(operations, modelOp)
				its.currentCheckPoint.Sseq = opDoc.Sseq
				return nil
			}); err != nil {
			return errors.NewPushPullError(errors.PushPullErrPullOperations, its.getPushPullTag(), err)
		}
		its.responsePushPullPack.Operations = operations
	}
	return nil
}

func (its *PushPullHandler) pushOperations() *errors.PushPullError {

	sseq := its.datatypeDoc.SseqEnd
	for _, op := range its.gotPushPullPack.Operations {
		// op := model.ToOperation(modelOp)
		if its.currentCheckPoint.Cseq+1 == op.ID.GetSeq() {
			sseq++
			// marshaledOp, err := proto.Marshal(op)
			// if err != nil {
			// 	return model.NewPushPullError(model.PushPullErrPushOperations, its.getPushPullTag(), err)
			// }
			opDoc := schema.NewOperationDoc(op, its.DUID, sseq, its.collectionDoc.Num)
			// opDoc := &schema.OperationDoc{
			// 	ID:            fmt.Sprintf("%s:%d", its.DUID, sseq),
			// 	DUID:          its.DUID,
			// 	CollectionNum: its.collectionDoc.Num,
			// 	OpType:        op.OpType.String(),
			// 	Sseq:          sseq,
			// 	// Operation:     string(marshaledOp),
			// 	CreatedAt:     time.Now(),
			// }
			its.pushingOperations = append(its.pushingOperations, opDoc)
			// its.responsePushPullPack.Operations = append(its.responsePushPullPack.Operations, modelOp)
			its.currentCheckPoint.SyncCseq(op.ID.GetSeq())
			its.currentCheckPoint.Sseq = sseq
		} else if its.currentCheckPoint.Cseq >= op.ID.GetSeq() {
			log.Logger.Warnf("reject operation due to duplicate: %v", op.String())
		} else {
			return errors.NewPushPullError(errors.PushPullErrMissingOperations, its.getPushPullTag(),
				fmt.Errorf("missing something in pushed operations: cp.Cseq=%d < op.Seq=%d",
					its.initialCheckPoint.Cseq, op.ID.GetSeq()))
		}
	}
	return nil
}

func (its *PushPullHandler) processSubscribeOrCreate(code pushPullCase) *errors.PushPullError {
	if its.gotOption.HasSubscribeBit() && its.gotOption.HasCreateBit() {
		switch code {
		case caseMatchNothing:
			its.createDatatype()
			return nil
		case caseAllMatchedNotSubscribed:
			return its.subscribeDatatype()
		}
	} else if its.gotOption.HasSubscribeBit() {
		switch code {
		case caseMatchNothing:
		case caseUsedDUID:
		case caseMatchKeyNotType:
		case caseAllMatchedSubscribed:
		case caseAllMatchedNotSubscribed:
			return its.subscribeDatatype()
		case caseAllMatchedNotVisible:
		}
	} else if its.gotOption.HasCreateBit() {
		switch code {
		case caseMatchNothing: // can create with key and duid
			its.createDatatype()
			return nil
		case caseUsedDUID: // duplicate DUID; can create with key but with another DUID
		case caseMatchKeyNotType: // key is already used;
		case caseAllMatchedSubscribed: // already created and subscribed; might duplicate creation; do nothing
		case caseAllMatchedNotSubscribed: // error: already created but not subscribed;
			return errors.NewPushPullError(errors.PushPullErrDuplicateDatatypeKey, its.getPushPullTag())
		case caseAllMatchedNotVisible: //

		default:

		}
	}
	return nil
}

func (its *PushPullHandler) subscribeDatatype() *errors.PushPullError {
	its.DUID = its.datatypeDoc.DUID
	its.clientDoc.CheckPoints[its.CUID] = its.currentCheckPoint
	its.gotPushPullPack.Operations = nil
	duid, err := types.DUIDFromString(its.datatypeDoc.DUID)
	if err != nil {
		return errors.NewPushPullError(errors.PushPullErrIllegalFormat, its.getPushPullTag(), "DUID", its.datatypeDoc.DUID)
	}
	its.responsePushPullPack.DUID = duid
	option := model.PushPullBitNormal
	its.responsePushPullPack.Option = uint32(*option.SetSubscribeBit())
	log.Logger.Infof("subscribe %s by %s", its.datatypeDoc, its.clientDoc)
	return nil
}

func (its *PushPullHandler) createDatatype() {
	its.datatypeDoc = &schema.DatatypeDoc{
		DUID:          its.DUID,
		Key:           its.Key,
		CollectionNum: its.collectionDoc.Num,
		Type:          model.TypeOfDatatype_name[its.gotPushPullPack.Type],
		SseqBegin:     1,
		SseqEnd:       0,
		Visible:       true,
		CreatedAt:     time.Now(),
	}
	option := model.PushPullBitNormal
	its.responsePushPullPack.Option = uint32(*option.SetCreateBit())
	log.Logger.Infof("create new %s", its.datatypeDoc)
}

func (its *PushPullHandler) evaluatePushPullCase() (pushPullCase, *errors.PushPullError) {
	var err error

	its.datatypeDoc, err = its.mongo.GetDatatypeByKey(its.ctx, its.collectionDoc.Num, its.gotPushPullPack.Key)
	if err != nil {
		return caseError, errors.NewPushPullError(errors.PushPullErrQueryToDB, its.getPushPullTag(), err)
	}
	if its.datatypeDoc == nil {
		its.datatypeDoc, err = its.mongo.GetDatatype(its.ctx, its.DUID)
		if err != nil {
			return caseError, errors.NewPushPullError(errors.PushPullErrQueryToDB, its.getPushPullTag(), err)
		}
		if its.datatypeDoc == nil {
			return caseMatchNothing, nil
		}
		return caseUsedDUID, nil
	}
	if its.datatypeDoc.Type == model.TypeOfDatatype_name[its.gotPushPullPack.Type] {
		if its.datatypeDoc.Visible {
			checkPoint := its.clientDoc.GetCheckPoint(its.DUID)
			if checkPoint != nil {
				its.initialCheckPoint = checkPoint
				its.currentCheckPoint = checkPoint.Clone()
				return caseAllMatchedSubscribed, nil
			}
			return caseAllMatchedNotSubscribed, nil
		}
		return caseAllMatchedNotVisible, nil
	}
	return caseMatchKeyNotType, nil

}
