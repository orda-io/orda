package service

import (
	"fmt"
	"github.com/orda-io/orda/client/pkg/context"
	errors2 "github.com/orda-io/orda/client/pkg/errors"
	model2 "github.com/orda-io/orda/client/pkg/model"
	"github.com/orda-io/orda/client/pkg/operations"
	"github.com/orda-io/orda/server/managers"
	"github.com/orda-io/orda/server/schema"
	"github.com/orda-io/orda/server/utils"
	"runtime/debug"
	"time"

	"github.com/orda-io/orda/server/constants"
	"github.com/orda-io/orda/server/snapshot"
	"github.com/orda-io/orda/server/svrcontext"
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

	err      errors2.OrdaError
	ctx      *svrcontext.ServerContext
	managers *managers.Managers
	lock     utils.Lock

	casePushPull      pushPullCase
	initialCheckPoint *model2.CheckPoint
	currentCP         *model2.CheckPoint

	clientDoc     *schema.ClientDoc
	datatypeDoc   *schema.DatatypeDoc
	collectionDoc *schema.CollectionDoc

	gotPushPullPack *model2.PushPullPack
	gotOption       *model2.PushPullPackOption

	resPushPullPack *model2.PushPullPack
	retCh           chan *model2.PushPullPack

	pushingOperations []interface{}
	pulledOperations  []model2.Operation
}

func newPushPullHandler(
	ctx *svrcontext.ServerContext,
	ppp *model2.PushPullPack,
	clientDoc *schema.ClientDoc,
	collectionDoc *schema.CollectionDoc,
	clients *managers.Managers,
) *PushPullHandler {

	newCtx := svrcontext.NewServerContext(ctx.Ctx(), constants.TagPushPull).
		UpdateCollection(collectionDoc.GetSummary()).
		UpdateClient(clientDoc.ToString()).UpdateDatatype(ppp.GetDatatypeTag())

	return &PushPullHandler{
		Key:             ppp.Key,
		DUID:            ppp.DUID,
		CUID:            clientDoc.CUID,
		ctx:             newCtx,
		err:             &errors2.MultipleOrdaErrors{},
		managers:        clients,
		clientDoc:       clientDoc,
		collectionDoc:   collectionDoc,
		gotPushPullPack: ppp,
		gotOption:       ppp.GetPushPullPackOption(),
	}
}

func (its *PushPullHandler) getLockKey() string {
	return fmt.Sprintf("PP:%d:%s", its.collectionDoc.Num, its.Key)
}

// Start begins the push-pull for a datatype and returns the result with the channel 'retCh'
func (its *PushPullHandler) Start() <-chan *model2.PushPullPack {
	retCh := make(chan *model2.PushPullPack)
	its.lock = its.managers.GetLock(its.ctx, its.getLockKey())
	go its.process(retCh)
	return retCh
}

func (its *PushPullHandler) initialize(retCh chan *model2.PushPullPack) errors2.OrdaError {
	its.retCh = retCh
	its.resPushPullPack = its.gotPushPullPack.GetResponsePushPullPack()
	its.resPushPullPack.Option = uint32(model2.PushPullBitNormal)

	checkPoint, err := its.managers.Mongo.GetCheckPointFromClient(its.ctx, its.CUID, its.DUID)
	if err != nil {
		return errors2.PushPullAbortionOfServer.New(its.ctx.L(), err.Error())
	}
	if checkPoint == nil {
		checkPoint = model2.NewCheckPoint()
	}

	its.initialCheckPoint = checkPoint.Clone()
	its.currentCP = checkPoint.Clone()
	its.ctx.L().Infof("REQ[PUPU] %v", its.gotPushPullPack.ToString())
	return nil
}

func (its *PushPullHandler) finalize() {
	if r := recover(); r != nil {
		its.ctx.L().Errorf("panic")
		return
	}
	defer its.lock.Unlock()
	if its.err == nil {
		its.ctx.L().Infof("finish with CP (%v) -> (%v) and pulled ops: %d",
			its.initialCheckPoint, its.currentCP, len(its.resPushPullPack.Operations))
		if len(its.pushingOperations) > 0 {

			newCtx := its.ctx.CloneWithNewContext(constants.TagPostPushPull)

			go func() {
				defer its.recoveryFromPanic()
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
	its.ctx.L().Infof("RES[PUPU] %s", its.resPushPullPack.ToString())
	its.retCh <- its.resPushPullPack
}

func (its *PushPullHandler) recoveryFromPanic() {
	if r := recover(); r != nil {
		its.ctx.L().Infof("finished from panic")
		debug.PrintStack()
		// TODO: need recovery process
	}
}

func (its *PushPullHandler) logInitialConditions() {
	its.ctx.L().Infof("initial condition | case: %s, opt: %s, cp(%v), ops:%d, sseqEnd:%d",
		pushPullCaseMap[its.casePushPull],
		its.gotOption.String(),
		its.initialCheckPoint,
		len(its.gotPushPullPack.Operations),
		its.datatypeDoc.SseqEnd,
	)
}

func (its *PushPullHandler) process(retCh chan *model2.PushPullPack) {

	its.lock.TryLock()

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
}

func (its *PushPullHandler) sendNotification(ctx context.OrdaContext) errors2.OrdaError {
	return its.managers.Notifier.NotifyAfterPushPull(
		ctx,
		its.collectionDoc.Name,
		its.clientDoc,
		its.datatypeDoc,
		its.currentCP.Sseq)
}

func (its *PushPullHandler) reserveUpdateSnapshot(ctx context.OrdaContext) error {
	snapshotManager := snapshot.NewManager(ctx, its.managers, its.datatypeDoc, its.collectionDoc)
	if err := snapshotManager.UpdateSnapshot(); err != nil { // TODO: should be asynchronous
		return err
	}
	return nil
}

func (its *PushPullHandler) commitToMongoDB() errors2.OrdaError {
	its.datatypeDoc.SseqEnd = its.currentCP.Sseq
	its.resPushPullPack.CheckPoint = its.currentCP

	if len(its.pushingOperations) > 0 {
		if err := its.managers.Mongo.InsertOperations(its.ctx, its.pushingOperations); err != nil {
			return errors2.PushPullAbortionOfServer.New(its.ctx.L(), err.Error())
		}
		its.ctx.L().Infof("commit %d Operations finally", len(its.pushingOperations))
	}

	if err := its.managers.Mongo.UpdateDatatype(its.ctx, its.datatypeDoc); err != nil {
		return errors2.PushPullAbortionOfServer.New(its.ctx.L(), err.Error())
	}
	its.ctx.L().Infof("commit Datatype %s", its.datatypeDoc)

	if !its.clientDoc.IsAdmin() {
		if err := its.managers.Mongo.UpdateCheckPointInClient(its.ctx, its.CUID, its.DUID, its.currentCP); err != nil {
			return errors2.PushPullAbortionOfServer.New(its.ctx.L(), err.Error())
		}
		its.ctx.L().Infof("commit CheckPoint with %s", its.currentCP.String())
	}
	return nil
}

func (its *PushPullHandler) pullOperations() errors2.OrdaError {
	if its.clientDoc.IsAdmin() {
		return nil
	}
	sseqBegin := its.gotPushPullPack.CheckPoint.Sseq + 1
	if its.datatypeDoc.SseqBegin <= sseqBegin && !its.gotOption.HasSnapshotBit() {
		opList, sseqList, err := its.managers.Mongo.GetOperations(its.ctx, its.DUID, sseqBegin, constants.InfinitySseq)
		if err != nil {
			return errors2.PushPullAbortionOfServer.New(its.ctx.L(), err.Error())
		}
		if len(opList) > 0 {
			its.currentCP.Sseq = sseqList[len(sseqList)-1] + (uint64)(len(its.pushingOperations))
		}
		its.resPushPullPack.Operations = opList
	}
	return nil
}

func (its *PushPullHandler) pushOperations() errors2.OrdaError {

	its.currentCP.Sseq = its.datatypeDoc.SseqEnd
	for _, op := range its.gotPushPullPack.Operations {
		switch {
		case its.currentCP.Cseq+1 == op.ID.GetSeq():
			its.currentCP.Sseq++
			opDoc := schema.NewOperationDoc(op, its.DUID, its.currentCP.Sseq, its.collectionDoc.Num)
			its.pushingOperations = append(its.pushingOperations, opDoc)
			its.ctx.L().Infof("%v) push %v", its.currentCP.Sseq, op.ToString())
			its.currentCP.SyncCseq(op.ID.GetSeq())
		case its.currentCP.Cseq >= op.ID.GetSeq():
			its.ctx.L().Warnf("reject operation due to duplicate: %v", op.String())
		default:
			msg := fmt.Sprintf("cp.Cseq=%d < op.Seq=%d", its.initialCheckPoint.Cseq, op.ID.GetSeq())
			return errors2.PushPullMissingOps.New(its.ctx.L(), msg)
		}
	}
	return nil
}

func (its *PushPullHandler) processSubscribeOrCreate(code pushPullCase) errors2.OrdaError {
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
			return errors2.PushPullNoDatatypeToSubscribe.New(its.ctx.L(), its.Key)
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
			return errors2.PushPullDuplicateKey.New(its.ctx.L(), its.Key)
		case caseAllMatchedNotVisible: //
		default:
		}
	}
	return nil
}

func (its *PushPullHandler) subscribeDatatype() errors2.OrdaError {
	its.DUID = its.datatypeDoc.DUID
	its.clientDoc.CheckPoints[its.CUID] = its.currentCP
	its.gotPushPullPack.Operations = nil
	its.resPushPullPack.DUID = its.datatypeDoc.DUID
	option := model2.PushPullBitNormal
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
	option := model2.PushPullBitNormal
	its.resPushPullPack.Option = uint32(*option.SetCreateBit())
	its.ctx.L().Infof("create new %s", its.datatypeDoc)
}

func (its *PushPullHandler) evaluatePushPullCase() (pushPullCase, errors2.OrdaError) {
	var err errors2.OrdaError
	its.datatypeDoc, err = its.managers.Mongo.GetDatatypeByKey(its.ctx, its.collectionDoc.Num, its.gotPushPullPack.Key)
	if err != nil {
		return caseError, err
	}
	if its.datatypeDoc == nil {
		its.datatypeDoc, err = its.managers.Mongo.GetDatatype(its.ctx, its.DUID)
		if err != nil {
			return caseError, errors2.PushPullAbortionOfServer.New(its.ctx.L(), err.Error())
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
				its.currentCP = checkPoint.Clone()
				return caseAllMatchedSubscribed, nil
			}
			return caseAllMatchedNotSubscribed, nil
		}
		return caseAllMatchedNotVisible, nil
	}
	return caseMatchKeyNotType, nil
}
