package managers

import (
	"github.com/knowhunger/ortoo/pkg/context"
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/model"
	"google.golang.org/grpc"
)

// SyncManager is a manager exchanging request and response with Ortoo server.
type SyncManager struct {
	seq                 uint32
	ctx                 context.OrtooContext
	conn                *grpc.ClientConn
	client              *model.Client
	serverAddr          string
	serviceClient       model.OrtooServiceClient
	notificationManager *NotificationManager
}

// NewSyncManager creates an instance of SyncManager.
func NewSyncManager(
	ctx context.OrtooContext,
	client *model.Client,
	serverAddr string,
	notificationAddr string,
) *SyncManager {
	var notificationManager *NotificationManager
	switch client.SyncType {
	case model.SyncType_LOCAL_ONLY, model.SyncType_MANUALLY:
		notificationManager = nil
	case model.SyncType_NOTIFIABLE:
		notificationManager = NewNotificationManager(ctx, notificationAddr, client)
	}
	return &SyncManager{
		seq:                 0,
		ctx:                 ctx,
		serverAddr:          serverAddr,
		client:              client,
		notificationManager: notificationManager,
	}
}

func (its *SyncManager) nextSeq() uint32 {
	currentSeq := its.seq
	its.seq++
	return currentSeq
}

// Connect makes connections with Ortoo GRPC and notification servers.
func (its *SyncManager) Connect() errors.OrtooError {
	conn, err := grpc.Dial(its.serverAddr, grpc.WithInsecure())
	if err != nil {
		return errors.ClientConnect.New(its.ctx.L(), err.Error())
	}
	its.conn = conn
	its.serviceClient = model.NewOrtooServiceClient(its.conn)
	its.ctx.L().Info("connect to grpc server")
	if its.notificationManager != nil {
		if err := its.notificationManager.Connect(); err != nil {
			return err
		}
	}
	return nil
}

// Close closes connections with Ortoo GRPC and notification servers.
func (its *SyncManager) Close() errors.OrtooError {

	if its.notificationManager != nil {
		its.notificationManager.Close()
	}
	if err := its.conn.Close(); err != nil {
		return errors.ClientClose.New(its.ctx.L(), err.Error())
	}
	return nil
}

// Sync exchanges PUSHPULL_REQUEST and PUSHPULL_RESPONSE
func (its *SyncManager) Sync(pppList ...*model.PushPullPack) (*model.PushPullResponse, errors.OrtooError) {
	request := model.NewPushPullRequest(its.nextSeq(), its.client, pppList...)
	its.ctx.L().Infof("SEND %s", request.ToString())
	response, err := its.serviceClient.ProcessPushPull(its.ctx, request)
	if err != nil {
		return nil, errors.ClientSync.New(its.ctx.L(), err.Error())
	}
	its.ctx.L().Infof("RECV %v", response.ToString())
	return response, nil
}

// ExchangeClientRequestResponse exchanges CLIENT_REQUEST and CLIENT_RESPONSE
func (its *SyncManager) ExchangeClientRequestResponse() errors.OrtooError {
	request := model.NewClientRequest(its.nextSeq(), its.client)
	its.ctx.L().Infof("SEND %s", request.ToString())
	response, err := its.serviceClient.ProcessClient(its.ctx, request)
	if err != nil {
		return errors.ClientSync.New(its.ctx.L(), err.Error())
	}
	its.ctx.L().Infof("RECV %s", response.ToString())
	return nil
}

func (its *SyncManager) subscribeNotification(topic string) errors.OrtooError {
	if its.notificationManager != nil {
		return its.notificationManager.SubscribeNotification(topic)
	}
	return nil
}

func (its *SyncManager) setNotificationReceiver(receiver notificationReceiver) {
	if its.notificationManager != nil {
		its.notificationManager.SetReceiver(receiver)
	}
}