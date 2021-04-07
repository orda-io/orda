package managers

import (
	"fmt"
	"github.com/knowhunger/ortoo/pkg/context"
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/iface"
	"github.com/knowhunger/ortoo/pkg/model"
	"strings"
)

// DatatypeManager manages Ortoo datatypes regarding operations
type DatatypeManager struct {
	ctx            context.OrtooContext
	cuid           string
	collectionName string
	syncManager    *SyncManager
	dataMap        map[string]iface.Datatype
}

// NewDatatypeManager creates a new instance of DatatypeManager
func NewDatatypeManager(
	ctx context.OrtooContext,
	client *model.Client,
	sm *SyncManager,
) *DatatypeManager {

	dm := &DatatypeManager{
		ctx:            ctx,
		cuid:           client.CUID,
		collectionName: client.Collection,
		dataMap:        make(map[string]iface.Datatype),
		syncManager:    sm,
	}
	if sm != nil {
		sm.setNotificationReceiver(dm)
	}
	return dm
}

// DeliverTransaction delivers a transaction
func (its *DatatypeManager) DeliverTransaction(wired iface.WiredDatatype) {
	// panic("implement me")
}

// ReceiveNotification enables datatype to sync when it receive notification
func (its *DatatypeManager) ReceiveNotification(topic string, notification model.NotificationPushPull) {
	if its.cuid == notification.CUID {
		its.ctx.L().Infof("drain own notification")
		return
	}
	splitTopic := strings.Split(topic, "/")
	datatypeKey := splitTopic[1]

	if err := its.SyncIfNeeded(datatypeKey, notification.DUID, notification.Sseq); err != nil {
		// _ = log.OrtooError(err)

	}

}

// OnChangeDatatypeState deals with what datatypeManager has to do when the state of datatype changes.
func (its *DatatypeManager) OnChangeDatatypeState(dt iface.Datatype, state model.StateOfDatatype) errors.OrtooError {
	switch state {
	case model.StateOfDatatype_SUBSCRIBED:
		topic := fmt.Sprintf("%s/%s", its.collectionName, dt.GetKey())
		if its.syncManager != nil {
			if err := its.syncManager.subscribeNotification(topic); err != nil {
				return errors.DatatypeSubscribe.New(nil, err.Error())
			}
			its.ctx.L().Infof("subscribe datatype topic(%s)", topic)
		}
	}
	return nil
}

// Get returns a datatype for the specified key
func (its *DatatypeManager) Get(key string) iface.Datatype {
	dt, ok := its.dataMap[key]
	if ok {
		return dt.GetDatatype()
	}
	return nil
}

// SubscribeOrCreate links a datatype with the datatype
func (its *DatatypeManager) SubscribeOrCreate(dt iface.Datatype, state model.StateOfDatatype) errors.OrtooError {
	if _, ok := its.dataMap[dt.GetKey()]; !ok {
		its.dataMap[dt.GetKey()] = dt
		if err := dt.SubscribeOrCreate(state); err != nil {
			return err
		}
	}
	return nil
}

// Sync enables a datatype of the specified key to be synchronized.
func (its *DatatypeManager) Sync(key string) errors.OrtooError {
	data := its.dataMap[key]
	ppp := data.CreatePushPullPack()
	pushPullResponse, err := its.syncManager.Sync(ppp)
	if err != nil {
		return err
	}
	for _, ppp := range pushPullResponse.PushPullPacks {
		if ppp.Key == data.GetKey() {
			data.ApplyPushPullPack(ppp)
		}
	}
	return nil
}

// SyncIfNeeded enables the datatype of the specified key and sseq to be synchronized if needed.
func (its *DatatypeManager) SyncIfNeeded(key, duid string, sseq uint64) errors.OrtooError {
	data, ok := its.dataMap[key]
	if ok {
		its.ctx.L().Infof("receive a notification for datatype %s(%s) sseq:%d", key, duid[0:8], sseq)
		if data.GetDUID() == duid && data.NeedSync(sseq) {
			its.ctx.L().Infof("need to sync after notification: %s (sseq:%d)", key, sseq)
			return its.Sync(key)
		}
	} else {
		its.ctx.L().Warnf("receive a notification for not subscribed datatype %s(%s) sseq:%d", key, duid, sseq)
	}
	return nil
}

// SyncAll enables all the subscribed datatypes to be synchronized.
func (its *DatatypeManager) SyncAll() errors.OrtooError {
	var pushPullPacks []*model.PushPullPack
	for _, data := range its.dataMap {
		ppp := data.CreatePushPullPack()
		pushPullPacks = append(pushPullPacks, ppp)
	}
	pushPullResponse, err := its.syncManager.Sync(pushPullPacks...)
	if err != nil {
		return err
	}
	for _, ppp := range pushPullResponse.PushPullPacks {
		if data, ok := its.dataMap[ppp.GetKey()]; ok {
			data.ApplyPushPullPack(ppp)
		}
	}
	return nil
}
