package managers

import (
	"github.com/knowhunger/ortoo/ortoo/context"
	"github.com/knowhunger/ortoo/ortoo/errors"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/ortoo/model"
	"google.golang.org/grpc"
)

// MessageManager is a manager exchanging request and response with Ortoo server.
type MessageManager struct {
	seq                 uint32
	host                string
	ctx                 *context.OrtooContext
	conn                *grpc.ClientConn
	client              *model.Client
	serviceClient       model.OrtooServiceClient
	notificationManager *NotificationManager
}

// NewMessageManager creates an instance of MessageManager.
func NewMessageManager(ctx *context.OrtooContext, client *model.Client, host string, notifyManager *NotificationManager) *MessageManager {

	return &MessageManager{
		seq:                 0,
		ctx:                 ctx,
		host:                host,
		client:              client,
		notificationManager: notifyManager,
	}
}

func (its *MessageManager) nextSeq() uint32 {
	currentSeq := its.seq
	its.seq++
	return currentSeq
}

// Connect makes connections with Ortoo GRPC and notification servers.
func (its *MessageManager) Connect() error {
	conn, err := grpc.Dial(its.host, grpc.WithInsecure())
	if err != nil {
		return errors.NewClientError(errors.ErrClientConnect, err.Error())
	}
	its.conn = conn
	its.serviceClient = model.NewOrtooServiceClient(its.conn)
	its.ctx.Logger.Info("connect to grpc server")
	if its.notificationManager != nil {
		if err := its.notificationManager.Connect(); err != nil {
			return err
		}
	}
	return nil
}

// Close closes connections with Ortoo GRPC and notification servers.
func (its *MessageManager) Close() error {

	if its.notificationManager != nil {
		its.notificationManager.Close()
	}
	if err := its.conn.Close(); err != nil {
		return errors.NewClientError(errors.ErrClientClose, err.Error())
	}
	return nil
}

// Sync exchanges PUSHPULL_REQUEST and PUSHPULL_RESPONSE
func (its *MessageManager) Sync(pppList ...*model.PushPullPack) (*model.PushPullResponse, error) {
	request := model.NewPushPullRequest(its.nextSeq(), its.client, pppList...)
	its.ctx.Logger.Infof("SEND %s", request.ToString())
	response, err := its.serviceClient.ProcessPushPull(its.ctx, request)
	if err != nil {
		return nil, log.OrtooErrorf(err, "fail to sync push pull")
	}
	its.ctx.Logger.Infof("RECV %v", response.ToString())
	return response, nil
}

// ExchangeClientRequestResponse exchanges CLIENT_REQUEST and CLIENT_RESPONSE
func (its *MessageManager) ExchangeClientRequestResponse() error {
	request := model.NewClientRequest(its.nextSeq(), its.client)
	its.ctx.Logger.Infof("SEND %s", request.ToString())
	response, err := its.serviceClient.ProcessClient(its.ctx, request)
	if err != nil {
		return log.OrtooErrorf(err, "fail to exchange clientRequestReply")
	}
	its.ctx.Logger.Infof("RECV %s", response.ToString())
	return nil
}
