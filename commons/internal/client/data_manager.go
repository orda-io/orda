package client

import (
	"github.com/knowhunger/ortoo/commons/internal/datatypes"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
)

//DataManager manages Ortoo datatypes regarding operations
type DataManager struct {
	reqResMgr *MessageManager
	dataMap   map[string]model.FinalDatatype
}

////DeliverOperation delivers an operation. It works differently according to delivery policy
//func (d *DataManager) DeliverOperation(wired datatypes.WiredDatatypeInterface, op model.Operation) {
//	//panic("implement me")
//}

//DeliverTransaction delivers a transaction
func (d *DataManager) DeliverTransaction(wired *datatypes.WiredDatatype) { //, transaction []model.Operation) {
	//panic("implement me")
}

func NewDataManager(manager *MessageManager) *DataManager {
	return &DataManager{
		dataMap:   make(map[string]model.FinalDatatype),
		reqResMgr: manager,
	}
}

//Get returns a datatype
func (d *DataManager) Get(key string) model.FinalDatatype {
	dt, ok := d.dataMap[key]
	if ok {
		return dt.GetFinalDatatype()
	}
	return nil
}

//SubscribeOrCreate links a datatype with the datatype
func (d *DataManager) SubscribeOrCreate(dt model.FinalDatatype, state model.StateOfDatatype) error {
	if _, ok := d.dataMap[dt.GetKey()]; !ok {
		d.dataMap[dt.GetKey()] = dt
		if err := dt.SubscribeOrCreate(state); err != nil {
			return log.OrtooErrorf(err, "fail to subscribe")
		}
	}
	return nil
}

func (d *DataManager) Sync(key string) error {
	data := d.dataMap[key]
	ppp := data.CreatePushPullPack()
	pushPullResponse, err := d.reqResMgr.Sync(ppp)
	if err != nil {
		return log.OrtooErrorf(err, "fail to sync")
	}
	for _, ppp := range pushPullResponse.PushPullPacks {
		if ppp.Key == data.GetKey() {
			data.ApplyPushPullPack(ppp)
		}
	}

	//log.Logger.Infof("%+v", pushPullResponse)
	return nil
}

func (d *DataManager) SyncAll() error {
	var pushPullPacks []*model.PushPullPack
	for _, data := range d.dataMap {
		ppp := data.CreatePushPullPack()
		pushPullPacks = append(pushPullPacks, ppp)
	}
	pushPullResponse, err := d.reqResMgr.Sync(pushPullPacks...)
	if err != nil {
		return log.OrtooError(err)
	}
	for _, ppp := range pushPullResponse.PushPullPacks {
		if data, ok := d.dataMap[ppp.GetKey()]; ok {
			data.ApplyPushPullPack(ppp)
		}
	}
	return nil
}
