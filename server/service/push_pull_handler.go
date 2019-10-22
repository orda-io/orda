package service

import (
	"context"
	"fmt"
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
	caseMatchKey
	caseMatchNothing
	caseMatchDUID
	caseMatchKeyNotType
	caseAllMatchedSubscribed
	caseAllMatchedNotSubscribed
	caseAllMatchedNotVisible
)

type PushPullHandler struct {
	ctx               context.Context
	checkPoint        *model.CheckPoint
	clientDoc         *schema.ClientDoc
	datatypeDoc       *schema.DatatypeDoc
	collectionDoc     *schema.CollectionDoc
	mongo             *mongodb.RepositoryMongo
	pushPullPack      *model.PushPullPack
	Option            model.PushPullPackOption
	pushingOperations []interface{}
	pulledOperations  []model.Operation
	DUID              string
	CUID              string
	Key               string
}

func (p *PushPullHandler) Start() <-chan *model.PushPullPack {
	retCh := make(chan *model.PushPullPack)
	go p.process(retCh)
	return retCh
}

func (p *PushPullHandler) process(retCh chan *model.PushPullPack) error {
	retPushPullPack := p.pushPullPack.GetReturnPushPullPack()

	checkPoint, err := p.mongo.GetCheckPointFromClient(p.ctx, p.CUID, p.DUID)
	if err != nil {
		_ = log.OrtooError(err)
		model.PushPullPackOption(retPushPullPack.Option).SetErrorBit()
		retCh <- retPushPullPack
		return
	}
	if checkPoint == nil {
		checkPoint = model.NewCheckPoint()
	}
	p.checkPoint = checkPoint
	casePushPull, err := p.evaluatePushPullCase()
	if err != nil {
		return log.OrtooError(err)
	}
	if err := p.processSubscribeOrCreate(casePushPull); err != nil {
		return log.OrtooError(err)
	}

	p.pushOperations()
	p.pullOperations()
	p.commitToMongoDB()

	return nil
}

func (p *PushPullHandler) commitToMongoDB() {

}

func (p *PushPullHandler) pullOperations() error {
	beginSseq := p.pushPullPack.CheckPoint.Sseq + 1
	cseq := p.checkPoint.GetCseq()
	operations, err := p.mongo.GetOperations(p.ctx, p.DUID, beginSseq, constants.InfinitySseq)
	if err != nil {
		return log.OrtooError(err)
	}

}

func (p *PushPullHandler) pushOperations() error {
	var operations []interface{}
	for _, opOnWire := range p.pushPullPack.Operations {
		op := model.ToOperation(opOnWire)
		if p.checkPoint.Cseq+1 == op.GetBase().ID.GetSeq() {
			operations = append(operations, op)
			p.checkPoint.SyncCseq(op.GetBase().ID.GetSeq())
		} else if p.checkPoint.Cseq > op.GetBase().ID.GetSeq() {
			log.Logger.Infof("reject operation due to duplicate: %v", op)
		} else {
			return log.OrtooError(
				fmt.Errorf("something is missing in pushed operations: checkpoint.Cseq=%d vs op.Seq=%d",
					p.checkPoint.Cseq, op.GetBase().ID.GetSeq()))
		}
	}
	p.pushingOperations = operations
	return nil
}

func (p *PushPullHandler) processSubscribeOrCreate(code pushPullCase) error {
	if p.Option.HasSubscribeBit() && p.Option.HasCreateBit() {

	} else if p.Option.HasSubscribeBit() {

	} else if p.Option.HasCreateBit() {
		switch code {
		case caseMatchNothing: // can create with key and duid
			if err := p.createDatatype(); err != nil {
				return log.OrtooError(err)
			}
		case caseMatchDUID: // duplicate DUID; can create with key but with another DUID
		case caseMatchKeyNotType: // key is already used;
		case caseAllMatchedSubscribed: // error: already created and subscribed; might duplicate creation
		case caseAllMatchedNotSubscribed: // error: already created but not subscribed;
		case caseAllMatchedNotVisible: //

		default:

		}
	}
	return nil
}

func (p *PushPullHandler) createDatatype() error {
	p.datatypeDoc = &schema.DatatypeDoc{
		DUID:          p.DUID,
		Key:           p.Key,
		CollectionNum: p.collectionDoc.Num,
		Type:          model.TypeOfDatatype_name[p.pushPullPack.Type],
		Sseq:          0,
		Visible:       true,
		CreatedAt:     time.Now(),
	}
	if err := p.mongo.UpdateDatatype(p.ctx, p.datatypeDoc); err != nil {
		return log.OrtooError(err)
	}
	return nil
}

func (p *PushPullHandler) evaluatePushPullCase() (pushPullCase, error) {
	var err error
	//if p.Option.HasSubscribeBit() {
	p.datatypeDoc, err = p.mongo.GetDatatypeByKey(p.ctx, p.collectionDoc.Num, p.pushPullPack.Key)
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
			return caseMatchDUID, nil
		}
	} else {
		if p.datatypeDoc.Type == model.TypeOfDatatype_name[p.pushPullPack.Type] {
			if p.datatypeDoc.Visible {
				checkPoint := p.clientDoc.GetCheckPoint(p.DUID)
				if checkPoint != nil {
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
