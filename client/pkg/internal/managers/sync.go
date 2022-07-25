package managers

import (
	"github.com/orda-io/orda/client/pkg/context"
	errors2 "github.com/orda-io/orda/client/pkg/errors"
	model2 "github.com/orda-io/orda/client/pkg/model"
	"google.golang.org/grpc"
)

// SyncManager is a manager exchanging request and response with Orda server.
type SyncManager struct {
	seq           uint32
	ctx           *context.ClientContext
	conn          *grpc.ClientConn
	client        *model2.Client
	serverAddr    string
	serviceClient model2.OrdaServiceClient
	notifyManager *NotifyManager
}

// NewSyncManager creates an instance of SyncManager.
func NewSyncManager(
	ctx *context.ClientContext,
	client *model2.Client,
	serverAddr string,
	notificationAddr string,
) *SyncManager {
	var notifyManager *NotifyManager
	switch client.SyncType {
	case model2.SyncType_LOCAL_ONLY, model2.SyncType_MANUALLY:
		notifyManager = nil
	case model2.SyncType_REALTIME:
		notifyManager = NewNotifyManager(ctx, notificationAddr, client)
	}
	return &SyncManager{
		seq:           0,
		ctx:           ctx,
		serverAddr:    serverAddr,
		client:        client,
		notifyManager: notifyManager,
	}
}

func (its *SyncManager) nextSeq() uint32 {
	currentSeq := its.seq
	its.seq++
	return currentSeq
}

// Connect makes connections with Orda GRPC and notification servers.
func (its *SyncManager) Connect() errors2.OrdaError {
	conn, err := grpc.Dial(its.serverAddr, grpc.WithInsecure())
	if err != nil {
		return errors2.ClientConnect.New(its.ctx.L(), err.Error())
	}
	its.conn = conn
	its.serviceClient = model2.NewOrdaServiceClient(its.conn)
	its.ctx.L().Info("connect to grpc server")
	if its.notifyManager != nil {
		if err := its.notifyManager.Connect(); err != nil {
			return err
		}
	}
	return nil
}

// Close closes connections with Orda GRPC and notification servers.
func (its *SyncManager) Close() errors2.OrdaError {
	if its.notifyManager != nil {
		its.notifyManager.Close()
	}
	if err := its.conn.Close(); err != nil {
		return errors2.ClientClose.New(its.ctx.L(), err.Error())
	}
	return nil
}

// Sync exchanges PUSHPULL_REQUEST and PUSHPULL_RESPONSE
func (its *SyncManager) Sync(pppList ...*model2.PushPullPack) (*model2.PushPullMessage, errors2.OrdaError) {
	request := model2.NewPushPullMessage(its.nextSeq(), its.client, pppList...)
	its.ctx.L().Infof("REQ[PUPU] %s", request.ToString())
	response, err := its.serviceClient.ProcessPushPull(its.ctx, request)
	if err != nil {
		return nil, errors2.ClientSync.New(its.ctx.L(), err.Error())
	}
	its.ctx.L().Infof("RES[PUPU] %v", response.ToString())
	return response, nil
}

// ExchangeClientRequestResponse exchanges CLIENT_REQUEST and CLIENT_RESPONSE
func (its *SyncManager) ExchangeClientRequestResponse() errors2.OrdaError {
	request := model2.NewClientMessage(its.client)

	response, err := its.serviceClient.ProcessClient(its.ctx, request)
	if err != nil {
		return errors2.ClientSync.New(its.ctx.L(), err.Error())
	}
	its.ctx.L().Infof("RES[CLIE] response: %s", response.ToString())
	return nil
}

func (its *SyncManager) subscribeNotification(topic string) errors2.OrdaError {
	if its.notifyManager != nil {
		return its.notifyManager.SubscribeNotification(topic)
	}
	return nil
}

func (its *SyncManager) setNotificationReceiver(receiver notificationReceiver) {
	if its.notifyManager != nil {
		its.notifyManager.SetReceiver(receiver)
	}
}
