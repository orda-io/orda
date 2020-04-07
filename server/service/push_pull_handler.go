package service

import (
	"context"
	"fmt"
	"github.com/knowhunger/ortoo/ortoo/errors"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/ortoo/model"
	"github.com/knowhunger/ortoo/ortoo/operations"
	"github.com/knowhunger/ortoo/ortoo/types"
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
func (p *PushPullHandler) Start() <-chan *model.PushPullPack {
	retCh := make(chan *model.PushPullPack)
	go p.process(retCh)
	return retCh
}

func (p *PushPullHandler) getPushPullTag() errors.PushPullTag {
	return errors.PushPullTag{
		CollectionName: p.collectionDoc.Name,
		Key:            p.Key,
		DUID:           p.DUID,
	}
}

func (p *PushPullHandler) initialize(retCh chan *model.PushPullPack) *errors.PushPullError {
	p.retCh = retCh
	p.responsePushPullPack = p.gotPushPullPack.GetResponsePushPullPack()

	checkPoint, err := p.mongo.GetCheckPointFromClient(p.ctx, p.CUID, p.DUID)
	if err != nil {
		return errors.NewPushPullError(errors.PushPullErrQueryToDB, p.getPushPullTag(), err)
	}
	if checkPoint == nil {
		checkPoint = model.NewCheckPoint()
	}

	p.initialCheckPoint = checkPoint.Clone()
	p.currentCheckPoint = checkPoint.Clone()
	return nil
}

func (p *PushPullHandler) finalize() {
	if p.err == nil {
		log.Logger.Infof("finish push-pull for key %s:%s (%v) -> (%v)", p.Key, p.DUID, p.initialCheckPoint, p.currentCheckPoint)
		if len(p.pushingOperations) > 0 {
			if err := p.sendNotification(); err == nil {
				// continue
			}
			if err := p.reserveUpdateSnapshot(); err != nil {
				// continue
			}
		}
	} else {
		p.responsePushPullPack.GetPushPullPackOption().SetErrorBit()

		errOp := operations.NewErrorOperation(p.err)
		p.responsePushPullPack.Operations = append(p.responsePushPullPack.Operations, errOp.ToModelOperation())

	}
	log.Logger.Infof("send back to %s %s", p.clientDoc.Alias, p.responsePushPullPack.ToString())
	p.retCh <- p.responsePushPullPack
}

func (p *PushPullHandler) logInitialConditions(casePushPull string) {
	log.Logger.Infof("show initial condition| case: %s, opt: %s, cp(%v), ops:%d",
		casePushPull, p.gotOption.String(), p.initialCheckPoint, len(p.gotPushPullPack.Operations))
}

func (p *PushPullHandler) process(retCh chan *model.PushPullPack) {
	defer p.finalize()

	if p.err = p.initialize(retCh); p.err != nil {
		return
	}

	casePushPull, err := p.evaluatePushPullCase()
	if p.err = err; err != nil {
		return
	}

	p.logInitialConditions(pushPullCaseMap[casePushPull])
	if p.err = p.processSubscribeOrCreate(casePushPull); p.err != nil {
		return
	}

	if p.err = p.pushOperations(); p.err != nil {
		return
	}
	if p.err = p.pullOperations(); p.err != nil {
		return
	}
	if p.err = p.commitToMongoDB(); p.err != nil {
		return
	}
	return
}

func (p *PushPullHandler) sendNotification() error {
	if err := p.notifier.NotifyAfterPushPull(
		p.collectionDoc.Name,
		p.datatypeDoc.Key,
		p.clientDoc.CUID,
		p.datatypeDoc.DUID,
		p.currentCheckPoint.Sseq); err != nil {
		return log.OrtooError(err)
	}

	return nil
}

func (p *PushPullHandler) reserveUpdateSnapshot() error {
	snapshotManager := snapshot.NewManager(p.ctx, p.mongo, p.datatypeDoc, p.collectionDoc)
	if err := snapshotManager.UpdateSnapshot(); err != nil { // TODO: should be asynchronous
		return log.OrtooError(err)
	}
	return nil
}

func (p *PushPullHandler) commitToMongoDB() *errors.PushPullError {
	p.datatypeDoc.SseqEnd = p.currentCheckPoint.Sseq
	p.responsePushPullPack.CheckPoint = p.currentCheckPoint

	if len(p.pushingOperations) > 0 {
		if err := p.mongo.InsertOperations(p.ctx, p.pushingOperations); err != nil {
			return errors.NewPushPullError(errors.PushPullErrQueryToDB, p.getPushPullTag(), err)
		}
		log.Logger.Infof("[MONGO] push %d operations finally", len(p.pushingOperations))
	}

	if err := p.mongo.UpdateDatatype(p.ctx, p.datatypeDoc); err != nil {
		return errors.NewPushPullError(errors.PushPullErrQueryToDB, p.getPushPullTag(), err)
	}
	log.Logger.Infof("[MONGO] update datatype with %+v", p.datatypeDoc)

	if err := p.mongo.UpdateCheckPointInClient(p.ctx, p.CUID, p.DUID, p.currentCheckPoint); err != nil {
		return errors.NewPushPullError(errors.PushPullErrQueryToDB, p.getPushPullTag(), err)
	}
	log.Logger.Infof("[MONGO] update '%s' with CP(%s)", p.clientDoc.Alias, p.currentCheckPoint.String())
	return nil
}

func (p *PushPullHandler) pullOperations() *errors.PushPullError {
	sseqBegin := p.gotPushPullPack.CheckPoint.Sseq + 1

	var operations []*model.Operation
	if p.datatypeDoc.SseqBegin <= sseqBegin && !p.gotOption.HasSnapshotBit() {
		if err := p.mongo.GetOperations(p.ctx,
			p.DUID,
			sseqBegin,
			constants.InfinitySseq,
			func(opDoc *schema.OperationDoc) error {
				var modelOp = opDoc.GetOperation()
				operations = append(operations, modelOp)
				p.currentCheckPoint.Sseq = opDoc.Sseq
				return nil
			}); err != nil {
			return errors.NewPushPullError(errors.PushPullErrPullOperations, p.getPushPullTag(), err)
		}
		p.responsePushPullPack.Operations = operations
	}
	return nil
}

func (p *PushPullHandler) pushOperations() *errors.PushPullError {

	sseq := p.initialCheckPoint.Sseq
	for _, op := range p.gotPushPullPack.Operations {
		// op := model.ToOperation(modelOp)
		if p.currentCheckPoint.Cseq+1 == op.ID.GetSeq() {
			sseq++
			// marshaledOp, err := proto.Marshal(op)
			// if err != nil {
			// 	return model.NewPushPullError(model.PushPullErrPushOperations, p.getPushPullTag(), err)
			// }
			opDoc := schema.NewOperationDoc(op, p.DUID, sseq, p.collectionDoc.Num)
			// opDoc := &schema.OperationDoc{
			// 	ID:            fmt.Sprintf("%s:%d", p.DUID, sseq),
			// 	DUID:          p.DUID,
			// 	CollectionNum: p.collectionDoc.Num,
			// 	OpType:        op.OpType.String(),
			// 	Sseq:          sseq,
			// 	// Operation:     string(marshaledOp),
			// 	CreatedAt:     time.Now(),
			// }
			p.pushingOperations = append(p.pushingOperations, opDoc)
			// p.responsePushPullPack.Operations = append(p.responsePushPullPack.Operations, modelOp)
			p.currentCheckPoint.SyncCseq(op.ID.GetSeq())
			p.currentCheckPoint.Sseq++
		} else if p.currentCheckPoint.Cseq >= op.ID.GetSeq() {
			log.Logger.Warnf("reject operation due to duplicate: %v", op.String())
		} else {
			return errors.NewPushPullError(errors.PushPullErrMissingOperations, p.getPushPullTag(),
				fmt.Errorf("missing something in pushed operations: cp.Cseq=%d < op.Seq=%d",
					p.initialCheckPoint.Cseq, op.ID.GetSeq()))
		}
	}
	return nil
}

func (p *PushPullHandler) processSubscribeOrCreate(code pushPullCase) *errors.PushPullError {
	if p.gotOption.HasSubscribeBit() && p.gotOption.HasCreateBit() {

	} else if p.gotOption.HasSubscribeBit() {
		switch code {
		case caseMatchNothing:
		case caseUsedDUID:
		case caseMatchKeyNotType:
		case caseAllMatchedSubscribed:
		case caseAllMatchedNotSubscribed:
			return p.subscribeDatatype()
		case caseAllMatchedNotVisible:
		}
	} else if p.gotOption.HasCreateBit() {
		switch code {
		case caseMatchNothing: // can create with key and duid
			p.setDatatype()
			return nil
		case caseUsedDUID: // duplicate DUID; can create with key but with another DUID
		case caseMatchKeyNotType: // key is already used;
		case caseAllMatchedSubscribed: // already created and subscribed; might duplicate creation; do nothing
		case caseAllMatchedNotSubscribed: // error: already created but not subscribed;
			return errors.NewPushPullError(errors.PushPullErrDuplicateDatatypeKey, p.getPushPullTag())
		case caseAllMatchedNotVisible: //

		default:

		}
	}
	return nil
}

func (p *PushPullHandler) subscribeDatatype() *errors.PushPullError {
	p.DUID = p.datatypeDoc.DUID
	p.clientDoc.CheckPoints[p.CUID] = p.currentCheckPoint
	p.gotPushPullPack.Operations = nil
	duid, err := types.DUIDFromString(p.datatypeDoc.DUID)
	if err != nil {
		return errors.NewPushPullError(errors.PushPullErrIllegalFormat, p.getPushPullTag(), "DUID", p.datatypeDoc.DUID)
	}
	p.responsePushPullPack.DUID = duid
	return nil
}

func (p *PushPullHandler) setDatatype() {
	p.datatypeDoc = &schema.DatatypeDoc{
		DUID:          p.DUID,
		Key:           p.Key,
		CollectionNum: p.collectionDoc.Num,
		Type:          model.TypeOfDatatype_name[p.gotPushPullPack.Type],
		SseqBegin:     1,
		SseqEnd:       0,
		Visible:       true,
		CreatedAt:     time.Now(),
	}
}

func (p *PushPullHandler) evaluatePushPullCase() (pushPullCase, *errors.PushPullError) {
	var err error

	p.datatypeDoc, err = p.mongo.GetDatatypeByKey(p.ctx, p.collectionDoc.Num, p.gotPushPullPack.Key)
	if err != nil {
		return caseError, errors.NewPushPullError(errors.PushPullErrQueryToDB, p.getPushPullTag(), err)
	}
	if p.datatypeDoc == nil {
		p.datatypeDoc, err = p.mongo.GetDatatype(p.ctx, p.DUID)
		if err != nil {
			return caseError, errors.NewPushPullError(errors.PushPullErrQueryToDB, p.getPushPullTag(), err)
		}
		if p.datatypeDoc == nil {
			return caseMatchNothing, nil
		}
		return caseUsedDUID, nil
	}
	if p.datatypeDoc.Type == model.TypeOfDatatype_name[p.gotPushPullPack.Type] {
		if p.datatypeDoc.Visible {
			checkPoint := p.clientDoc.GetCheckPoint(p.DUID)
			if checkPoint != nil {
				p.initialCheckPoint = checkPoint
				p.currentCheckPoint = checkPoint.Clone()
				return caseAllMatchedSubscribed, nil
			}
			return caseAllMatchedNotSubscribed, nil
		}
		return caseAllMatchedNotVisible, nil
	}
	return caseMatchKeyNotType, nil

}
