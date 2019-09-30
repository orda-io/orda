package client

import (
	"github.com/knowhunger/ortoo/commons/internal/datatypes"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
)

//DataManager manages Ortoo datatypes regarding operations
type DataManager struct {
	//cuid model.Cuid
	reqResMgr *MessageManager
	dataMap   map[string]model.FinalDatatype
}

//DeliverOperation delivers an operation
func (d *DataManager) DeliverOperation(wired datatypes.WiredDatatype, op model.Operation) {
	//panic("implement me")
}

//DeliverTransaction delivers a transaction
func (d *DataManager) DeliverTransaction(wired datatypes.WiredDatatype, transaction []model.Operation) {
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

//Subscribe links a datatype with the datatype
func (d *DataManager) Subscribe(dt model.FinalDatatype) error {
	if _, ok := d.dataMap[dt.GetKey()]; !ok {
		d.dataMap[dt.GetKey()] = dt
		if err := dt.Subscribe(); err != nil {
			return log.OrtooErrorf(err, "fail to subscribe")
		}
	}
	return nil
}

func (d *DataManager) Sync(key string) error {
	data := d.dataMap[key]
	ppp := data.CreatePushPullPack()
	pushpullResponse, err := d.reqResMgr.Sync(ppp)
	if err != nil {
		return log.OrtooErrorf(err, "fail to sync")
	}
	log.Logger.Infof("%+v", pushpullResponse)
	return nil
}
