package service

import (
	gocontext "context"
	"fmt"
	"github.com/knowhunger/ortoo/pkg/context"
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/pkg/operations"
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

	err      errors.OrtooError
	ctx      context.OrtooContext
	mongo    *mongodb.RepositoryMongo
	notifier *notification.Notifier

	casePushPull      pushPullCase
	initialCheckPoint *model.CheckPoint
	currentCheckPoint *model.CheckPoint

	clientDoc     *schema.ClientDoc
	datatypeDoc   *schema.DatatypeDoc
	collectionDoc *schema.CollectionDoc

	gotPushPullPack *model.PushPullPack
	gotOption       *model.PushPullPackOption

	resPushPullPack *model.PushPullPack
	retCh           chan *model.PushPullPack

	pushingOperations []interface{}
	pulledOperations  []model.Operation
}

func newPushPullHandler(
	ctx gocontext.Context,
	ppp *model.PushPullPack,
	clientDoc *schema.ClientDoc,
	collectionDoc *schema.CollectionDoc,
	service *OrtooService,
) *PushPullHandler {
	ortooCtx := context.NewWithTags(ctx, context.SERVER,
		context.MakeTagInPushPull(constants.TagPushPull, collectionDoc.Num, clientDoc.CUID, ppp.DUID))
	return &PushPullHandler{
		Key:             ppp.Key,
		DUID:            ppp.DUID,
		CUID:            clientDoc.CUID,
		ctx:             ortooCtx,
		err:             &errors.MultipleOrtooErrors{},
		mongo:           service.mongo,
		notifier:        service.notifier,
		clientDoc:       clientDoc,
		collectionDoc:   collectionDoc,
		gotPushPullPack: ppp,
		gotOption:       ppp.GetPushPullPackOption(),
	}
}

// Start begins the push pull for a datatype and returns the result with the channel 'retCh'
func (its *PushPullHandler) Start() <-chan *model.PushPullPack {
	retCh := make(chan *model.PushPullPack)
	go its.process(retCh)
	return retCh
}

func (its *PushPullHandler) initialize(retCh chan *model.PushPullPack) errors.OrtooError {
	its.retCh = retCh
	its.resPushPullPack = its.gotPushPullPack.GetResponsePushPullPack()
	its.resPushPullPack.Option = uint32(model.PushPullBitNormal)

	checkPoint, err := its.mongo.GetCheckPointFromClient(its.ctx, its.CUID, its.DUID)
	if err != nil {
		return errors.PushPullAbortedForServer.New(its.ctx.L(), err.Error())
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
		its.ctx.L().Infof("finish with CP (%v) -> (%v) and pulled ops: %d",
			its.initialCheckPoint, its.currentCheckPoint, len(its.resPushPullPack.Operations))
		if len(its.pushingOperations) > 0 {
			newCtx := context.NewWithTags(gocontext.TODO(), context.SERVER,
				context.MakeTagInPushPull(constants.TagPostPushPull, its.collectionDoc.Num, its.clientDoc.CUID, its.resPushPullPack.DUID))
			go func() {
				if err := its.sendNotification(newCtx); err == nil {
					// continue
				}
				if err := its.reserveUpdateSnapshot(newCtx); err != nil {
					// continue
				}
			}()
		}
	} else {
		its.ctx.L().Infof("finish with an error: %v", its.err.Error())
		its.resPushPullPack.GetPushPullPackOption().SetErrorBit()
		errOp := operations.NewErrorOperation(its.err)
		its.resPushPullPack.Operations = append(its.resPushPullPack.Operations, errOp.ToModelOperation())
	}
	its.ctx.L().Infof("SENDBACK %s", its.resPushPullPack.ToString())
	its.retCh <- its.resPushPullPack
}

func (its *PushPullHandler) logInitialConditions() {
	its.ctx.L().Infof("initial condition| case: %s, opt: %s, cp(%v), ops:%d, sseqEnd:%d",
		pushPullCaseMap[its.casePushPull],
		its.gotOption.String(),
		its.initialCheckPoint,
		len(its.gotPushPullPack.Operations),
		its.datatypeDoc.SseqEnd,
	)
}

func (its *PushPullHandler) process(retCh chan *model.PushPullPack) {
	defer its.finalize()

	if its.err = its.initialize(retCh); its.err != nil {
		return
	}

	if its.casePushPull, its.err = its.evaluatePushPullCase(); its.err != nil {
		return
	}

	if its.err = its.processSubscribeOrCreate(its.casePushPull); its.err != nil {
		return
	}

	its.logInitialConditions()

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

func (its *PushPullHandler) sendNotification(ctx context.OrtooContext) errors.OrtooError {
	if err := its.notifier.NotifyAfterPushPull(
		ctx,
		its.collectionDoc.Name,
		its.clientDoc,
		its.datatypeDoc,
		its.currentCheckPoint.Sseq); err != nil {
		return err
	}
	return nil
}

func (its *PushPullHandler) reserveUpdateSnapshot(ctx context.OrtooContext) error {
	snapshotManager := snapshot.NewManager(ctx, its.mongo, its.datatypeDoc, its.collectionDoc)
	if err := snapshotManager.UpdateSnapshot(); err != nil { // TODO: should be asynchronous
		return err
	}
	return nil
}

func (its *PushPullHandler) commitToMongoDB() errors.OrtooError {
	its.datatypeDoc.SseqEnd = its.currentCheckPoint.Sseq
	its.resPushPullPack.CheckPoint = its.currentCheckPoint

	if len(its.pushingOperations) > 0 {
		if err := its.mongo.InsertOperations(its.ctx, its.pushingOperations); err != nil {
			return errors.PushPullAbortedForServer.New(its.ctx.L(), err.Error())
		}
		its.ctx.L().Infof("commit %d Operations finally", len(its.pushingOperations))
	}

	if err := its.mongo.UpdateDatatype(its.ctx, its.datatypeDoc); err != nil {
		return errors.PushPullAbortedForServer.New(its.ctx.L(), err.Error())
	}
	its.ctx.L().Infof("commit Datatype %s", its.datatypeDoc)

	if err := its.mongo.UpdateCheckPointInClient(its.ctx, its.CUID, its.DUID, its.currentCheckPoint); err != nil {
		return errors.PushPullAbortedForServer.New(its.ctx.L(), err.Error())
	}
	its.ctx.L().Infof("commit CheckPoint with %s", its.currentCheckPoint.String())
	return nil
}

func (its *PushPullHandler) pullOperations() errors.OrtooError {
	sseqBegin := its.gotPushPullPack.CheckPoint.Sseq + 1
	if its.datatypeDoc.SseqBegin <= sseqBegin && !its.gotOption.HasSnapshotBit() {
		opList, sseqList, err := its.mongo.GetOperations(its.ctx, its.DUID, sseqBegin, constants.InfinitySseq)
		if err != nil {
			return errors.PushPullAbortedForServer.New(its.ctx.L(), err.Error())
		}
		if len(opList) > 0 {
			its.currentCheckPoint.Sseq = sseqList[len(sseqList)-1]
		}
		its.resPushPullPack.Operations = opList
	}
	return nil
}

func (its *PushPullHandler) pushOperations() errors.OrtooError {

	sseq := its.datatypeDoc.SseqEnd
	for _, op := range its.gotPushPullPack.Operations {
		// op := model.ToOperation(modelOp)
		if its.currentCheckPoint.Cseq+1 == op.ID.GetSeq() {
			sseq++
			// marshaledOp, err := proto.Marshal(op)
			// if err != nil {
			// 	return model.NewPushPullError(model.PushPullPushOperations, its.getPushPullTag(), err)
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
			// its.resPushPullPack.Operations = append(its.resPushPullPack.Operations, modelOp)
			its.currentCheckPoint.SyncCseq(op.ID.GetSeq())
			its.currentCheckPoint.Sseq = sseq
		} else if its.currentCheckPoint.Cseq >= op.ID.GetSeq() {
			its.ctx.L().Warnf("reject operation due to duplicate: %v", op.String())
		} else {
			msg := fmt.Sprintf("cp.Cseq=%d < op.Seq=%d", its.initialCheckPoint.Cseq, op.ID.GetSeq())
			return errors.PushPullMissingOps.New(its.ctx.L(), msg)
		}
	}
	return nil
}

func (its *PushPullHandler) processSubscribeOrCreate(code pushPullCase) errors.OrtooError {
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
			return errors.PushPullDuplicateKey.New(its.ctx.L(), its.Key)
		case caseAllMatchedNotVisible: //
		default:
		}
	}
	return nil
}

func (its *PushPullHandler) subscribeDatatype() errors.OrtooError {
	its.DUID = its.datatypeDoc.DUID
	its.clientDoc.CheckPoints[its.CUID] = its.currentCheckPoint
	its.gotPushPullPack.Operations = nil
	its.resPushPullPack.DUID = its.datatypeDoc.DUID
	option := model.PushPullBitNormal
	its.resPushPullPack.Option = uint32(*option.SetSubscribeBit())
	its.ctx.L().Infof("subscribe %s by %s", its.datatypeDoc, its.clientDoc)
	return nil
}

func (its *PushPullHandler) createDatatype() {
	its.datatypeDoc = &schema.DatatypeDoc{
		DUID:          its.DUID,
		Key:           its.Key,
		CollectionNum: its.collectionDoc.Num,
		Type:          its.gotPushPullPack.Type.String(), // model.TypeOfDatatype_name[its.gotPushPullPack.Type],
		SseqBegin:     1,
		SseqEnd:       0,
		Visible:       true,
		CreatedAt:     time.Now(),
	}
	option := model.PushPullBitNormal
	its.resPushPullPack.Option = uint32(*option.SetCreateBit())
	its.ctx.L().Infof("create new %s", its.datatypeDoc)
}

func (its *PushPullHandler) evaluatePushPullCase() (pushPullCase, errors.OrtooError) {
	var err errors.OrtooError
	its.datatypeDoc, err = its.mongo.GetDatatypeByKey(its.ctx, its.collectionDoc.Num, its.gotPushPullPack.Key)
	if err != nil {
		return caseError, err
	}
	if its.datatypeDoc == nil {
		its.datatypeDoc, err = its.mongo.GetDatatype(its.ctx, its.DUID)
		if err != nil {
			return caseError, errors.PushPullAbortedForServer.New(its.ctx.L(), err.Error())
		}
		if its.datatypeDoc == nil {
			return caseMatchNothing, nil
		}
		return caseUsedDUID, nil
	}
	if its.datatypeDoc.Type == its.gotPushPullPack.Type.String() {
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
