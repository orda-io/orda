package service

import (
	"context"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
	"github.com/knowhunger/ortoo/server/constants"
	"github.com/knowhunger/ortoo/server/mongodb"
	"github.com/knowhunger/ortoo/server/mongodb/schema"
	"time"
)

type pushPullCase uint32

const (
	caseError pushPullCase = iota
	//caseMatchKey
	caseMatchNothing
	caseUsedDUID
	caseMatchKeyNotType
	caseAllMatchedSubscribed
	caseAllMatchedNotSubscribed
	caseAllMatchedNotVisible
)

type PushPullHandler struct {
	Key  string
	DUID string
	CUID string

	ctx   context.Context
	mongo *mongodb.RepositoryMongo

	initialCheckPoint *model.CheckPoint
	currentCheckPoint *model.CheckPoint

	clientDoc     *schema.ClientDoc
	datatypeDoc   *schema.DatatypeDoc
	collectionDoc *schema.CollectionDoc

	gotPushPullPack *model.PushPullPack
	gotOption       *model.PushPullPackOption

	retPushPullPack *model.PushPullPack
	retCh           chan *model.PushPullPack

	pushingOperations []interface{}
	pulledOperations  []model.Operation
}

func (p *PushPullHandler) Start() <-chan *model.PushPullPack {
	retCh := make(chan *model.PushPullPack)
	go p.process(retCh)
	return retCh
}

func (p *PushPullHandler) finalize(err error) {
	p.retCh <- p.retPushPullPack
	log.Logger.Infof("processed push-pull:%s", p.DUID)
}

func (p *PushPullHandler) initialize(retCh chan *model.PushPullPack) error {
	p.retCh = retCh
	p.retPushPullPack = p.gotPushPullPack.GetReturnPushPullPack()

	checkPoint, err := p.mongo.GetCheckPointFromClient(p.ctx, p.CUID, p.DUID)
	if err != nil {
		p.retPushPullPack.GetPushPullPackOption().SetErrorBit()
		return log.OrtooError(err)
	}
	if checkPoint == nil {
		checkPoint = model.NewCheckPoint()
	}

	p.initialCheckPoint = checkPoint.Clone()
	p.currentCheckPoint = checkPoint.Clone()
	return nil
}

func (p *PushPullHandler) process(retCh chan *model.PushPullPack) (err error) {
	defer p.finalize(err)

	if err := p.initialize(retCh); err != nil {
		return log.OrtooError(err)
	}

	casePushPull, err := p.evaluatePushPullCase()
	if err != nil {
		return log.OrtooError(err)
	}
	log.Logger.Infof("PushPullCase: %d", casePushPull)
	if err := p.processSubscribeOrCreate(casePushPull); err != nil {
		return log.OrtooError(err)
	}

	if err := p.pushOperations(); err != nil {
		return log.OrtooError(err)
	}
	if err := p.pullOperations(); err != nil {
		return log.OrtooError(err)
	}
	if err := p.commitToMongoDB(); err != nil {
		return log.OrtooError(err)
	}

	return nil
}

func (p *PushPullHandler) commitToMongoDB() error {
	p.datatypeDoc.SseqEnd = p.currentCheckPoint.Sseq
	p.retPushPullPack.CheckPoint = p.currentCheckPoint
	if err := p.mongo.InsertOperations(p.ctx, p.pushingOperations); err != nil {
		return log.OrtooError(err)
	}

	if err := p.mongo.UpdateDatatype(p.ctx, p.datatypeDoc); err != nil {
		return log.OrtooError(err)
	}

	if err := p.mongo.UpdateCheckPointInClient(p.ctx, p.CUID, p.DUID, p.currentCheckPoint); err != nil {
		return log.OrtooError(err)
	}

	return nil
}

func (p *PushPullHandler) pullOperations() error {
	sseqBegin := p.gotPushPullPack.CheckPoint.Sseq + 1

	var operations []*model.OperationOnWire
	if p.datatypeDoc.SseqBegin <= sseqBegin && !p.gotOption.HasSnapshotBit() {
		err := p.mongo.GetOperations(p.ctx, p.DUID, sseqBegin, constants.InfinitySseq, func(opDoc *schema.OperationDoc) error {
			var opOnWire model.OperationOnWire
			if err := proto.Unmarshal(opDoc.Operation, &opOnWire); err != nil {
				_ = log.OrtooError(err)
				return nil
			}
			operations = append(operations, &opOnWire)
			p.currentCheckPoint.Sseq = opDoc.Sseq
			return nil
		})
		if err != nil {
			return log.OrtooError(err)
		}
		p.retPushPullPack.Operations = operations

	}
	return nil
}

func (p *PushPullHandler) pushOperations() error {

	sseq := p.initialCheckPoint.Sseq
	for _, opOnWire := range p.gotPushPullPack.Operations {
		op := model.ToOperation(opOnWire)
		if p.currentCheckPoint.Cseq+1 == op.GetBase().ID.GetSeq() {
			sseq++
			marshaledOp, err := proto.Marshal(opOnWire)
			if err != nil {
				return log.OrtooError(err)
			}
			opDoc := &schema.OperationDoc{
				ID:            fmt.Sprintf("%s:%d", p.DUID, sseq),
				DUID:          p.DUID,
				CollectionNum: p.collectionDoc.Num,
				OpType:        op.GetBase().OpType.String(),
				Sseq:          sseq,
				Operation:     marshaledOp,
				CreatedAt:     time.Now(),
			}
			p.pushingOperations = append(p.pushingOperations, opDoc)
			p.retPushPullPack.Operations = append(p.retPushPullPack.Operations, opOnWire)
			p.currentCheckPoint.SyncCseq(op.GetBase().ID.GetSeq())
			p.currentCheckPoint.Sseq++

		} else if p.currentCheckPoint.Cseq >= op.GetBase().ID.GetSeq() {
			log.Logger.Infof("reject operation due to duplicate: %v", op)
		} else {
			return log.OrtooError(
				fmt.Errorf("something is missing in pushed operations: checkpoint.Cseq=%d vs op.Seq=%d",
					p.initialCheckPoint.Cseq, op.GetBase().ID.GetSeq()))
		}
	}
	return nil
}

func (p *PushPullHandler) processSubscribeOrCreate(code pushPullCase) error {
	if p.gotOption.HasSubscribeBit() && p.gotOption.HasCreateBit() {

	} else if p.gotOption.HasSubscribeBit() {
		switch code {
		case caseMatchNothing:
		case caseUsedDUID:
		case caseMatchKeyNotType:
		case caseAllMatchedSubscribed:
		case caseAllMatchedNotSubscribed:
			if err := p.subscribeDatatype(); err != nil {
				return log.OrtooError(err)
			}
		case caseAllMatchedNotVisible:
		}
	} else if p.gotOption.HasCreateBit() {
		switch code {
		case caseMatchNothing: // can create with key and duid
			if err := p.createDatatype(); err != nil {
				return log.OrtooError(err)
			}
		case caseUsedDUID: // duplicate DUID; can create with key but with another DUID
		case caseMatchKeyNotType: // key is already used;
		case caseAllMatchedSubscribed: // already created and subscribed; might duplicate creation; do nothing
		case caseAllMatchedNotSubscribed: // error: already created but not subscribed;
		case caseAllMatchedNotVisible: //

		default:

		}
	}
	return nil
}

func (p *PushPullHandler) subscribeDatatype() error {
	p.DUID = p.datatypeDoc.DUID
	p.clientDoc.CheckPoints[p.CUID] = p.currentCheckPoint
	p.gotPushPullPack.Operations = nil

	return nil
}

func (p *PushPullHandler) createDatatype() error {
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
	//if err := p.mongo.UpdateDatatype(p.ctx, p.datatypeDoc); err != nil {
	//	return log.OrtooError(err)
	//}
	return nil
}

func (p *PushPullHandler) evaluatePushPullCase() (pushPullCase, error) {
	var err error

	p.datatypeDoc, err = p.mongo.GetDatatypeByKey(p.ctx, p.collectionDoc.Num, p.gotPushPullPack.Key)
	if err != nil {
		return caseError, log.OrtooError(err)
	}
	if p.datatypeDoc == nil {
		p.datatypeDoc, err = p.mongo.GetDatatype(p.ctx, p.DUID)
		if err != nil {
			return caseError, log.OrtooError(err)
		}
		if p.datatypeDoc == nil {
			return caseMatchNothing, nil
		} else {
			return caseUsedDUID, nil
		}
	} else {
		if p.datatypeDoc.Type == model.TypeOfDatatype_name[p.gotPushPullPack.Type] {
			if p.datatypeDoc.Visible {
				checkPoint := p.clientDoc.GetCheckPoint(p.DUID)
				if checkPoint != nil {
					p.initialCheckPoint = checkPoint
					p.currentCheckPoint = checkPoint.Clone()
					return caseAllMatchedSubscribed, nil
				} else {
					return caseAllMatchedNotSubscribed, nil
				}
			} else {
				return caseAllMatchedNotVisible, nil
			}
		} else {
			return caseMatchKeyNotType, nil
		}
	}

}
