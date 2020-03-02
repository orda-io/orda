package managers

import (
	"encoding/hex"
	"fmt"
	"github.com/knowhunger/ortoo/commons/context"
	"github.com/knowhunger/ortoo/commons/errors"
	"github.com/knowhunger/ortoo/commons/internal/datatypes"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
	"strings"
)

// DatatypeManager manages Ortoo datatypes regarding operations
type DatatypeManager struct {
	ctx                 *context.OrtooContext
	cuid                string
	collectionName      string
	messageManager      *MessageManager
	notificationManager *NotificationManager
	dataMap             map[string]model.CommonDatatype
}

// DeliverTransaction delivers a transaction
func (d *DatatypeManager) DeliverTransaction(wired *datatypes.WiredDatatype) {
	// panic("implement me")
}

// ReceiveNotification enables datatype to sync when it receive notification
func (d *DatatypeManager) ReceiveNotification(topic string, notification model.NotificationPushPull) {
	if d.cuid == notification.CUID {
		d.ctx.Logger.Infof("drain own notification")
		return
	}
	splitTopic := strings.Split(topic, "/")
	datatypeKey := splitTopic[1]
	if err := d.SyncIfNeeded(datatypeKey, notification.DUID, notification.Sseq); err != nil {
		_ = log.OrtooError(err)
	}
}

// OnChangeDatatypeState deals with what datatypeManager has to do when the state of datatype changes.
func (d *DatatypeManager) OnChangeDatatypeState(dt model.CommonDatatype, state model.StateOfDatatype) error {
	switch state {
	case model.StateOfDatatype_SUBSCRIBED:
		topic := fmt.Sprintf("%s/%s", d.collectionName, dt.GetKey())
		if d.notificationManager != nil {
			if err := d.notificationManager.SubscribeNotification(topic); err != nil {
				return errors.NewDatatypeError(errors.ErrDatatypeSubscribe, err.Error())
			}
			d.ctx.Logger.Infof("subscribe datatype topic: %s", topic)
		}
	}
	return nil
}

// NewDatatypeManager creates a new instance of DatatypeManager
func NewDatatypeManager(ctx *context.OrtooContext, mm *MessageManager, nm *NotificationManager, collectionName string, cuid model.CUID) *DatatypeManager {
	dm := &DatatypeManager{
		ctx:                 ctx,
		cuid:                hex.EncodeToString(cuid),
		collectionName:      collectionName,
		dataMap:             make(map[string]model.CommonDatatype),
		messageManager:      mm,
		notificationManager: nm,
	}
	mm.SetNotificationReceiver(dm)
	return dm
}

// Get returns a datatype for the given key
func (d *DatatypeManager) Get(key string) model.CommonDatatype {
	dt, ok := d.dataMap[key]
	if ok {
		return dt.GetFinalDatatype()
	}
	return nil
}

// SubscribeOrCreate links a datatype with the datatype
func (d *DatatypeManager) SubscribeOrCreate(dt model.CommonDatatype, state model.StateOfDatatype) error {
	if _, ok := d.dataMap[dt.GetKey()]; !ok {
		d.dataMap[dt.GetKey()] = dt
		if err := dt.SubscribeOrCreate(state); err != nil {
			return log.OrtooErrorf(err, "fail to subscribe")
		}
	}
	return nil
}

// Sync enables a datatype of the given key to be synchronized.
func (d *DatatypeManager) Sync(key string) error {
	data := d.dataMap[key]
	ppp := data.CreatePushPullPack()
	pushPullResponse, err := d.messageManager.Sync(ppp)
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

// SyncIfNeeded enables a datatype of the given key and sseq to be synchronized if needed.
func (d *DatatypeManager) SyncIfNeeded(key, duid string, sseq uint64) error {
	data, ok := d.dataMap[key]
	if ok {
		d.ctx.Logger.Infof("receive a notification for datatype %s(%s) sseq:%d", key, duid, sseq)
		if hex.EncodeToString(data.GetDUID()) == duid && data.NeedSync(sseq) {
			d.ctx.Logger.Infof("need to sync after notification: %s (sseq:%d)", key, sseq)
			return d.Sync(key)
		}
	} else {
		d.ctx.Logger.Warnf("receive a notification for not subscribed datatype %s(%s) sseq:%d", key, duid, sseq)
	}

	return nil
}

// SyncAll enables all the subscribed datatypes to be synchronized.
func (d *DatatypeManager) SyncAll() error {
	var pushPullPacks []*model.PushPullPack
	for _, data := range d.dataMap {
		ppp := data.CreatePushPullPack()
		pushPullPacks = append(pushPullPacks, ppp)
	}
	pushPullResponse, err := d.messageManager.Sync(pushPullPacks...)
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
