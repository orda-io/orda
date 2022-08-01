package managers

import (
	"fmt"
	"github.com/orda-io/orda/client/pkg/context"
	"github.com/orda-io/orda/client/pkg/errors"
	"github.com/orda-io/orda/client/pkg/iface"
	"github.com/orda-io/orda/client/pkg/model"
	"strings"

	"golang.org/x/sync/semaphore"
)

// DatatypeManager manages Orda datatypes regarding operations
type DatatypeManager struct {
	ctx         *context.ClientContext
	syncManager *SyncManager
	sema        *semaphore.Weighted
	dataMap     map[string]iface.Datatype
}

// NewDatatypeManager creates a new instance of DatatypeManager
func NewDatatypeManager(ctx *context.ClientContext, sm *SyncManager) *DatatypeManager {
	dm := &DatatypeManager{
		ctx:         ctx,
		dataMap:     make(map[string]iface.Datatype),
		syncManager: sm,
		sema:        semaphore.NewWeighted(1),
	}
	if sm != nil {
		sm.setNotificationReceiver(dm)
	}
	return dm
}

// DeliverTransaction delivers a transaction
func (its *DatatypeManager) DeliverTransaction(wired iface.WiredDatatype) {
	if its.ctx.Client.SyncType == model.SyncType_REALTIME {
		go func() {
			if !its.sema.TryAcquire(1) {

				return
			}
			defer func() {
				its.sema.Release(1)
				if wired.NeedPush() {
					its.ctx.L().Infof("deliver transaction after delivering")
					its.DeliverTransaction(wired)
				}
			}()
			if err := its.sync(wired); err != nil {
				// TODO: handle in ErrorHandler
			}
		}()
	}
}

func (its *DatatypeManager) ExistDatatype(key string, typeOf model.TypeOfDatatype) (iface.Datatype, errors.OrdaError) {
	if data, ok := its.dataMap[key]; ok {
		if data.GetType() == typeOf {
			its.ctx.L().Warnf("already subscribed datatype '%s'", key)
			return data, nil
		}
		err := errors.DatatypeSubscribe.New(nil,
			fmt.Sprintf("not matched type: %s vs %s", typeOf.String(), data.GetType().String()))
		return nil, err
	}
	return nil, nil
}

// ReceiveNotification enables datatype to sync when it receive notification
func (its *DatatypeManager) ReceiveNotification(topic string, notification model.Notification) {
	if its.ctx.Client.CUID == notification.CUID {
		its.ctx.L().Infof("drain own notification")
		return
	}
	splitTopic := strings.Split(topic, "/")
	datatypeKey := splitTopic[1]
	if data, ok := its.dataMap[datatypeKey]; ok && data.GetDUID() == notification.DUID {
		if err := its.syncIfNeedPull(data, notification.Sseq); err != nil {
			// TODO: call errorHandler
			return
		}
		return
	}
	its.ctx.L().Warnf(
		"receive a notification for not subscribed datatype %s(%s) sseq:%d",
		datatypeKey,
		notification.DUID,
		notification.Sseq,
	)
}

// SyncAll enables all the subscribed datatypes to be synchronized.
func (its *DatatypeManager) SyncAll() errors.OrdaError {
	if err := its.sema.Acquire(its.ctx.Ctx(), 1); err != nil {
		return errors.ClientSync.New(its.ctx.L())
	}
	defer func() {
		its.sema.Release(1)
	}()

	var pushPullPacks []*model.PushPullPack
	for _, data := range its.dataMap {
		ppp := data.CreatePushPullPack()
		pushPullPacks = append(pushPullPacks, ppp)
	}
	return its.syncPushPullPacks(pushPullPacks...)
}

// syncIfNeedPull enables the datatype of the specified key and sseq to be synchronized if needed.
func (its *DatatypeManager) syncIfNeedPull(data iface.WiredDatatype, sseq uint64) errors.OrdaError {
	if data.NeedPull(sseq) {
		its.ctx.L().Infof("need to sync after notification: %s (sseq:%d)", data.GetKey(), sseq)
		return its.sync(data)
	}
	return nil
}

// OnChangeDatatypeState deals with what datatypeManager has to do when the state of datatype changes.
func (its *DatatypeManager) OnChangeDatatypeState(dt iface.Datatype, state model.StateOfDatatype) errors.OrdaError {
	if state == model.StateOfDatatype_SUBSCRIBED {
		topic := fmt.Sprintf("%s/%s", its.ctx.Client.Collection, dt.GetKey())
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
		return dt
	}
	return nil
}

// SubscribeOrCreate links a datatype with the datatype
func (its *DatatypeManager) SubscribeOrCreate(dt iface.Datatype, state model.StateOfDatatype) errors.OrdaError {
	if _, ok := its.dataMap[dt.GetKey()]; !ok {
		its.dataMap[dt.GetKey()] = dt
		if err := dt.SubscribeOrCreate(state); err != nil {
			return err
		}
	}
	return nil
}

// sync enables a datatype of the specified key to be synchronized.
func (its *DatatypeManager) sync(data iface.WiredDatatype) errors.OrdaError {
	ppp := data.CreatePushPullPack()
	return its.syncPushPullPacks(ppp)
}

func (its *DatatypeManager) needPush() bool {
	for _, data := range its.dataMap {
		if data.NeedPush() {
			return true
		}
	}
	return false
}

func (its *DatatypeManager) syncPushPullPacks(pppList ...*model.PushPullPack) errors.OrdaError {
	pushPullResponse, err := its.syncManager.Sync(pppList...)
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
