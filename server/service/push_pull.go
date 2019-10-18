package service

import (
	"context"
	"encoding/hex"
	"github.com/golang/protobuf/proto"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
	"github.com/knowhunger/ortoo/server/mongodb"
	"github.com/knowhunger/ortoo/server/mongodb/schema"
	"reflect"
)

//ProcessPushPull processes a GRPC for Push-Pull
func (o *OrtooService) ProcessPushPull(ctx context.Context, in *model.PushPullRequest) (*model.PushPullResponse, error) {
	log.Logger.Infof("Received: %v, %s", proto.MarshalTextString(in), hex.EncodeToString(in.Header.GetCuid()))
	collectionDoc, err := o.mongo.GetCollection(ctx, in.Header.GetCollection())
	if collectionDoc == nil || err != nil {
		return nil, model.NewRPCError(model.RPCErrMongoDB)
	}

	clientDoc, err := o.mongo.GetClient(ctx, hex.EncodeToString(in.Header.GetCuid()))
	if err != nil {
		return nil, model.NewRPCError(model.RPCErrMongoDB)
	}
	if clientDoc == nil {
		return nil, model.NewRPCError(model.RPCErrNoClient)
	}
	if clientDoc.CollectionNum != collectionDoc.Num {
		return nil, model.NewRPCError(model.RPCErrClientInconsistentCollection, clientDoc.CollectionNum, collectionDoc.Name)
	}
	var chanList []<-chan *model.PushPullPack
	for _, ppp := range in.PushPullPacks {
		handler := &PushPullHandler{
			ctx:           ctx,
			clientDoc:     clientDoc,
			collectionDoc: collectionDoc,
			mongo:         o.mongo,
			pushPullPack:  ppp,
			Option:        model.PushPullPackOption(ppp.Option),
			DUID:          hex.EncodeToString(ppp.DUID),
			CUID:          clientDoc.CUID,
			Key:           ppp.Key,
		}
		chanList = append(chanList, handler.Start())
	}
	remainingChan := len(chanList)
	cases := make([]reflect.SelectCase, remainingChan)
	for i, ch := range chanList {
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
	}
	for remainingChan > 0 {
		chosen, value, ok := reflect.Select(cases)
		if !ok {
			_ = log.OrtooErrorf(nil, "fail to run")
		}
		ch := chanList[chosen]
		msg := value.Interface()

		log.Logger.Infof("%v %v", ch, msg)
	}

	return &model.PushPullResponse{Id: in.Id}, nil
}

func NewPushPullHandler(
	ctx context.Context,
	mongo *mongodb.RepositoryMongo,
	ppp *model.PushPullPack,
	collectionDoc *schema.CollectionDoc,
	clientDoc *schema.ClientDoc) *PushPullHandler {
	return &PushPullHandler{
		ctx:           ctx,
		collectionDoc: collectionDoc,
		clientDoc:     clientDoc,
		mongo:         mongo,
		pushPullPack:  ppp,
	}
}

type PushPullHandler struct {
	ctx           context.Context
	checkPoint    *model.CheckPoint
	clientDoc     *schema.ClientDoc
	datatypeDoc   *schema.DatatypeDoc
	collectionDoc *schema.CollectionDoc
	mongo         *mongodb.RepositoryMongo
	pushPullPack  *model.PushPullPack
	Option        model.PushPullPackOption
	DUID          string
	CUID          string
	Key           string
}

func (p *PushPullHandler) Start() <-chan *model.PushPullPack {
	retCh := make(chan *model.PushPullPack)
	go p.process(retCh)
	return retCh
}

func (p *PushPullHandler) process(retCh chan *model.PushPullPack) {
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

	//retCh <-
}

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

func (p *PushPullHandler) evaluate(code pushPullCase) {
	if p.Option.HasSubscribeBit() && p.Option.HasCreateBit() {

	} else if p.Option.HasSubscribeBit() {

	} else if p.Option.HasCreateBit() {

	}
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
	//} else if p.Option {
	//	p.datatypeDoc, err = p.mongo.GetDatatype(p.ctx, p.DUID)
	//	if p.datatypeDoc == nil {
	//		p.datatypeDoc, err = p.mongo.GetDatatypeByKey(p.ctx, p.collectionDoc.Num, p.Key)
	//		if p.datatypeDoc == nil {
	//			return caseMatchNothing, nil
	//		} else {
	//			return caseMatchKey, nil
	//		}
	//	} else {
	//		if p.Key == p.datatypeDoc.Key {
	//
	//		} else {
	//
	//		}
	//	}
	//}

}
