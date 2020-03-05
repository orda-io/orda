package managers

import (
	"github.com/knowhunger/ortoo/commons/context"
	"github.com/knowhunger/ortoo/commons/errors"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
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

func (m *MessageManager) nextSeq() uint32 {
	currentSeq := m.seq
	m.seq++
	return currentSeq
}

// Connect makes connections with Ortoo GRPC and notification servers.
func (m *MessageManager) Connect() error {
	conn, err := grpc.Dial(m.host, grpc.WithInsecure())
	if err != nil {
		return errors.NewClientError(errors.ErrClientConnect, err.Error())
	}
	m.conn = conn
	m.serviceClient = model.NewOrtooServiceClient(m.conn)
	m.ctx.Logger.Info("connect to grpc server")
	if m.notificationManager != nil {
		if err := m.notificationManager.Connect(); err != nil {
			return err
		}
	}
	return nil
}

// Close closes connections with Ortoo GRPC and notification servers.
func (m *MessageManager) Close() error {
	if err := m.conn.Close(); err != nil {
		return errors.NewClientError(errors.ErrClientClose, err.Error())
	}
	if m.notificationManager != nil {
		m.notificationManager.Close()
	}
	return nil
}

// Sync exchanges PUSHPULL_REQUEST and PUSHPULL_RESPONSE
func (m *MessageManager) Sync(pppList ...*model.PushPullPack) (*model.PushPullResponse, error) {
	request := model.NewPushPullRequest(m.nextSeq(), m.client, pppList...)
	m.ctx.Logger.Infof("send PUSHPULL REQUEST:%s", request.ToString())
	response, err := m.serviceClient.ProcessPushPull(m.ctx, request)
	if err != nil {
		return nil, log.OrtooErrorf(err, "fail to sync push pull")
	}
	m.ctx.Logger.Infof("receive PUSHPULL RESPONSE:%v", response.ToString())
	return response, nil
}

// ExchangeClientRequestResponse exchanges CLIENT_REQUEST and CLIENT_RESPONSE
func (m *MessageManager) ExchangeClientRequestResponse() error {
	request := model.NewClientRequest(m.nextSeq(), m.client)
	m.ctx.Logger.Infof("send CLIENT REQUEST:%s", request.ToString())
	response, err := m.serviceClient.ProcessClient(m.ctx, request)
	if err != nil {
		return log.OrtooErrorf(err, "fail to exchange clientRequestReply")
	}
	m.ctx.Logger.Infof("receive CLIENT RESPONSE: %s", response.ToString())
	return nil
}
