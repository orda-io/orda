package client

import (
	"encoding/hex"
	"fmt"
	"github.com/knowhunger/ortoo/commons/internal/datatypes"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
	"strings"
)

// DataManager manages Ortoo datatypes regarding operations
type DataManager struct {
	cuid           string
	collectionName string
	msgMgr         *MessageManager
	dataMap        map[string]model.FinalDatatype
}

// DeliverTransaction delivers a transaction
func (d *DataManager) DeliverTransaction(wired *datatypes.WiredDatatype) {
	// panic("implement me")
}

func (d *DataManager) ReceiveNotification(topic string, notification model.NotificationPushPull) {
	if d.cuid == notification.CUID {
		log.Logger.Infof("drain own notification")
		return
	}
	splitTopic := strings.Split(topic, "/")
	datatypeKey := splitTopic[1]
	if err := d.SyncIfNeeded(datatypeKey, notification.DUID, notification.Sseq); err != nil {
		_ = log.OrtooError(err)
	}
}

func (d *DataManager) OnChangeDatatypeState(dt model.FinalDatatype, state model.StateOfDatatype) {
	switch state {
	case model.StateOfDatatype_SUBSCRIBED:
		topic := fmt.Sprintf("%s/%s", d.collectionName, dt.GetKey())
		d.msgMgr.SubscribePubSub(topic)
		dt.HandleSubscription()
	}
}

func NewDataManager(manager *MessageManager, collectionName string, cuid model.CUID) *DataManager {

	dm := &DataManager{
		cuid:           hex.EncodeToString(cuid),
		collectionName: collectionName,
		dataMap:        make(map[string]model.FinalDatatype),
		msgMgr:         manager,
	}
	manager.SetNotificationReceiver(dm)
	return dm
}

// Get returns a datatype
func (d *DataManager) Get(key string) model.FinalDatatype {
	dt, ok := d.dataMap[key]
	if ok {
		return dt.GetFinalDatatype()
	}
	return nil
}

// SubscribeOrCreate links a datatype with the datatype
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
	pushPullResponse, err := d.msgMgr.Sync(ppp)
	if err != nil {
		return log.OrtooErrorf(err, "fail to sync")
	}
	for _, ppp := range pushPullResponse.PushPullPacks {
		if ppp.Key == data.GetKey() {
			data.ApplyPushPullPack(ppp)
		}
	}
	return nil
}

func (d *DataManager) SyncIfNeeded(key, duid string, sseq uint64) error {
	data := d.dataMap[key]
	if hex.EncodeToString(data.GetDUID()) == duid && data.NeedSync(sseq) {
		log.Logger.Infof("need to sync after notification: %s (sseq:%d)", key, sseq)
		return d.Sync(key)
	}
	return nil
}

func (d *DataManager) SyncAll() error {
	var pushPullPacks []*model.PushPullPack
	for _, data := range d.dataMap {
		ppp := data.CreatePushPullPack()
		pushPullPacks = append(pushPullPacks, ppp)
	}
	pushPullResponse, err := d.msgMgr.Sync(pushPullPacks...)
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
