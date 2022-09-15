package service

import (
	"fmt"
	"github.com/orda-io/orda/client/pkg/errors"
	"github.com/orda-io/orda/client/pkg/iface"
	"github.com/orda-io/orda/client/pkg/model"
	"github.com/orda-io/orda/client/pkg/operations"
	"github.com/orda-io/orda/server/constants"
	"github.com/orda-io/orda/server/managers"
	"github.com/orda-io/orda/server/schema"
	"github.com/orda-io/orda/server/snapshot"
	"github.com/orda-io/orda/server/svrcontext"
	"github.com/orda-io/orda/server/utils"
	"runtime/debug"
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
	caseInvalidPushPullOption
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
		caseInvalidPushPullOption:   "caseInvalidPushPullOption",
	}
)

// PushPullHandler is a struct that handles a push-pull for a datatype
type PushPullHandler struct {
	Key  string
	DUID string
	CUID string

	err      errors.OrdaError
	ctx      *svrcontext.ServerContext
	managers *managers.Managers
	lock     utils.Lock

	casePushPull pushPullCase
	initialCP    *model.CheckPoint
	currentCP    *model.CheckPoint

	datatypeDoc   *schema.DatatypeDoc
	clientDoc     *schema.ClientDoc
	subClientDoc  *schema.SubscribedClientDoc
	collectionDoc *schema.CollectionDoc

	gotPushPullPack *model.PushPullPack
	gotOption       *model.PushPullPackOption
	isReadOnly      bool

	resPushPullPack *model.PushPullPack
	retCh           chan *model.PushPullPack

	pushingOperations []interface{}
	pulledOperations  []model.Operation
}

func newPushPullHandler(
	ctx *svrcontext.ServerContext,
	ppp *model.PushPullPack,
	clientDoc *schema.ClientDoc,
	collectionDoc *schema.CollectionDoc,
	clients *managers.Managers,
) *PushPullHandler {

	newCtx := svrcontext.NewServerContext(ctx.Ctx(), constants.TagPushPull).
		UpdateCollection(collectionDoc.GetSummary()).
		UpdateClient(clientDoc.ToString()).UpdateDatatype(ppp.GetDatatypeTag())
	option := ppp.GetPushPullPackOption()
	return &PushPullHandler{
		Key:             ppp.Key,
		DUID:            ppp.DUID,
		CUID:            clientDoc.CUID,
		ctx:             newCtx,
		err:             &errors.MultipleOrdaErrors{},
		managers:        clients,
		collectionDoc:   collectionDoc,
		clientDoc:       clientDoc,
		gotPushPullPack: ppp,
		gotOption:       option,
		isReadOnly:      option.HasReadOnly(),
	}
}

func (its *PushPullHandler) getLockKey() string {
	return fmt.Sprintf("PP:%d:%s", its.collectionDoc.Num, its.Key)
}

// Start begins the push-pull for a datatype and returns the result with the channel 'retCh'
func (its *PushPullHandler) Start() <-chan *model.PushPullPack {
	retCh := make(chan *model.PushPullPack)
	its.lock = its.managers.GetLock(its.ctx, its.getLockKey())
	go its.process(retCh)
	return retCh
}

func (its *PushPullHandler) validatePushPullPack() errors.OrdaError {
	if its.isReadOnly && its.gotOption.HasCreateBit() {
		return errors.PushPullAbortionOfClient.New(its.ctx.L(), "invalid push-pull option:"+its.gotOption.String())
	}
	if its.isReadOnly && len(its.gotPushPullPack.Operations) > 0 {
		return errors.PushPullAbortionOfClient.New(its.ctx.L(), "the readonly client cannot push operations")
	}
	return nil
}

func (its *PushPullHandler) initialize(retCh chan *model.PushPullPack) errors.OrdaError {
	its.retCh = retCh
	its.resPushPullPack = its.gotPushPullPack.GetResponsePushPullPack()
	its.resPushPullPack.Option = uint32(model.PushPullBitNormal)

	its.ctx.L().Infof("REQ[PUPU] %v", its.gotPushPullPack.ToString(true))
	return nil
}

func (its *PushPullHandler) finalize() {
	if r := recover(); r != nil {
		its.ctx.L().Errorf("recover panic [%v]: %v", r, string(debug.Stack()))

		return
	}
	defer its.lock.Unlock()
	if its.err == nil {
		its.ctx.L().Infof("finish with CP %v -> %v and pulled ops: %d",
			its.initialCP.ToString(), its.currentCP.ToString(), len(its.resPushPullPack.Operations))
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
	its.ctx.L().Infof("RES[PUPU] %s", its.resPushPullPack.ToString(true))
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
	its.ctx.L().Infof("initial condition | case: %s, opt: %s, cp%v, ops:%d, sseqEnd:%d",
		pushPullCaseMap[its.casePushPull],
		its.gotOption.String(),
		its.initialCP.ToString(),
		len(its.gotPushPullPack.Operations),
		its.datatypeDoc.Sseq.End,
	)
}

func (its *PushPullHandler) process(retCh chan *model.PushPullPack) {

	its.lock.TryLock()

	defer its.finalize()

	if its.err = its.validatePushPullPack(); its.err != nil {
		return
	}

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

func (its *PushPullHandler) sendNotification(ctx iface.OrdaContext) errors.OrdaError {
	return its.managers.Notifier.NotifyAfterPushPull(
		ctx,
		its.collectionDoc.Name,
		its.CUID,
		its.datatypeDoc,
		its.currentCP.Sseq)
}

func (its *PushPullHandler) reserveUpdateSnapshot(ctx iface.OrdaContext) error {
	snapshotManager := snapshot.NewManager(ctx, its.managers, its.datatypeDoc, its.collectionDoc)
	if err := snapshotManager.UpdateSnapshot(); err != nil { // TODO: should be asynchronous
		return err
	}
	return nil
}

func (its *PushPullHandler) commitToMongoDB() errors.OrdaError {
	its.datatypeDoc.Sseq.End = its.currentCP.Sseq
	its.resPushPullPack.CheckPoint = its.currentCP
	its.subClientDoc.UpdateAt()
	if len(its.pushingOperations) > 0 {
		if err := its.managers.Mongo.InsertOperations(its.ctx, its.pushingOperations); err != nil {
			return errors.PushPullAbortionOfServer.New(its.ctx.L(), err.Error())
		}
		its.ctx.L().Infof("commit %d OperationDocs", len(its.pushingOperations))
	}

	if err := its.managers.Mongo.UpdateDatatype(its.ctx, its.datatypeDoc); err != nil {
		return errors.PushPullAbortionOfServer.New(its.ctx.L(), err.Error())
	}
	its.ctx.L().Infof("commit DatatypeDoc [%s]", its.datatypeDoc)

	// if !admin.IsAdminCUID(its.CUID) {
	// 	if err := its.managers.Mongo.UpdateCheckPointInClient(its.ctx, its.CUID, its.DUID, its.currentCP); err != nil {
	// 		return errors.PushPullAbortionOfServer.New(its.ctx.L(), err.Error())
	// 	}
	// 	its.ctx.L().Infof("commit CheckPoint with %s", its.currentCP.String())
	// }
	return nil
}

func (its *PushPullHandler) pullOperations() errors.OrdaError {
	if its.clientDoc.GetType() == model.ClientType_VOLATILE {
		return nil
	}
	sseqBegin := its.gotPushPullPack.CheckPoint.Sseq + 1
	if its.datatypeDoc.Sseq.Begin <= sseqBegin && !its.gotOption.HasSnapshotBit() {
		opList, sseqList, err := its.managers.Mongo.GetOperations(its.ctx, its.DUID, sseqBegin, constants.InfinitySseq)
		if err != nil {
			return errors.PushPullAbortionOfServer.New(its.ctx.L(), err.Error())
		}
		if len(opList) > 0 {
			its.currentCP.Sseq = sseqList[len(sseqList)-1] + (uint64)(len(its.pushingOperations))
		}
		its.resPushPullPack.Operations = opList
	}
	return nil
}

func (its *PushPullHandler) pushOperations() errors.OrdaError {
	if its.isReadOnly {
		return nil
	}
	its.currentCP.Sseq = its.datatypeDoc.Sseq.End
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
			msg := fmt.Sprintf("cp.Cseq=%d < op.Seq=%d", its.initialCP.Cseq, op.ID.GetSeq())
			return errors.PushPullMissingOps.New(its.ctx.L(), msg)
		}
	}
	return nil
}

func (its *PushPullHandler) processSubscribeOrCreate(code pushPullCase) errors.OrdaError {
	if its.gotOption.HasSubscribeBit() && its.gotOption.HasCreateBit() {
		switch code {
		case caseMatchNothing:
			return its.createDatatype()
		case caseAllMatchedNotSubscribed:
			return its.subscribeDatatype()
		}
	} else if its.gotOption.HasSubscribeBit() {
		switch code {
		case caseMatchNothing:
			return errors.PushPullNoDatatypeToSubscribe.New(its.ctx.L(), its.Key)
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
			return its.createDatatype()
		case caseUsedDUID: // duplicate DUID; can create with key but with another DUID
		case caseMatchKeyNotType: // key is already used;
		case caseAllMatchedSubscribed: // already created and subscribed; might duplicate creation; do nothing
		case caseAllMatchedNotSubscribed: // error: already created but not subscribed;
			return errors.PushPullDuplicateKey.New(its.ctx.L(), its.Key)
		case caseAllMatchedNotVisible: //
		default:
		}
	}
	return its.initClientInfoWithDatatypeDoc()
}

func (its *PushPullHandler) subscribeDatatype() errors.OrdaError {
	its.DUID = its.datatypeDoc.DUID
	if err := its.initClientInfoWithDatatypeDoc(); err != nil {
		return err
	}
	its.gotPushPullPack.Operations = nil
	its.resPushPullPack.DUID = its.datatypeDoc.DUID
	option := model.PushPullBitNormal
	its.resPushPullPack.Option = uint32(*option.SetSubscribeBit())
	its.ctx.L().Infof("subscribe %v", its.datatypeDoc)
	return nil
}

func (its *PushPullHandler) createDatatype() errors.OrdaError {
	its.datatypeDoc = schema.NewDatatypeDoc(its.DUID, its.Key, its.collectionDoc.Num, its.gotPushPullPack.Type.String())
	option := model.PushPullBitNormal
	its.resPushPullPack.Option = uint32(*option.SetCreateBit())
	its.ctx.L().Infof("create new %v", its.datatypeDoc)
	if err := its.initClientInfoWithDatatypeDoc(); err != nil {
		return err
	}
	return nil
}

func (its *PushPullHandler) initClientInfoWithDatatypeDoc() errors.OrdaError {
	// if its.cli
	its.subClientDoc = its.datatypeDoc.GetClientInDatatypeDoc(its.CUID, its.isReadOnly)
	if its.subClientDoc != nil {
		its.currentCP = its.subClientDoc.GetCheckPoint()
		its.initialCP = its.currentCP.Clone()
		return nil
	}
	if its.clientDoc.GetType() != model.ClientType_VOLATILE {
		its.subClientDoc = its.datatypeDoc.AddNewClient(its.CUID, its.clientDoc.Type, its.isReadOnly)
		its.currentCP = its.subClientDoc.GetCheckPoint()
	} else {
		its.currentCP = model.NewCheckPoint()
	}
	its.initialCP = its.currentCP.Clone()
	return nil
}

func (its *PushPullHandler) evaluatePushPullCase() (pushPullCase, errors.OrdaError) {
	var err errors.OrdaError
	// (1) first, when having createBit or subscribeBit, check if there exists a datatypeDoc for the key
	if its.gotOption.HasCreateBit() || its.gotOption.HasSubscribeBit() {
		its.datatypeDoc, err = its.managers.Mongo.GetDatatypeByKey(its.ctx, its.collectionDoc.Num, its.gotPushPullPack.Key)
		if err != nil {
			return caseError, errors.PushPullAbortionOfServer.New(its.ctx.L(), "fail to get datatype by key from DB")
		}
	}

	// (2) except with createBit or subscribeBit, search datatypeDoc by DUID
	if its.datatypeDoc == nil {
		its.datatypeDoc, err = its.managers.Mongo.GetDatatype(its.ctx, its.DUID)
		if err != nil {
			return caseError, errors.PushPullAbortionOfServer.New(its.ctx.L(), "fail to get datatype by duid from DB")
		}
		if its.datatypeDoc == nil {
			return caseMatchNothing, nil
		}
		return caseUsedDUID, nil
	}
	if its.datatypeDoc.Type == its.gotPushPullPack.Type.String() {
		if its.datatypeDoc.Visible {
			subscribedClient := its.datatypeDoc.GetClientInDatatypeDoc(its.CUID, its.isReadOnly)
			if subscribedClient != nil {
				return caseAllMatchedSubscribed, nil
			}
			return caseAllMatchedNotSubscribed, nil
		}
		return caseAllMatchedNotVisible, nil
	}
	return caseMatchKeyNotType, nil
}
