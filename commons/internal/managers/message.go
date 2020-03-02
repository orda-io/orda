package managers

import (
	"github.com/knowhunger/ortoo/commons/context"
	"github.com/knowhunger/ortoo/commons/errors"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
	"google.golang.org/grpc"
)

// MessageManager is a manager exchanging request and response.
type MessageManager struct {
	seq                 uint32
	host                string
	ctx                 *context.OrtooContext
	conn                *grpc.ClientConn
	client              *model.Client
	serviceClient       model.OrtooServiceClient
	notificationManager *NotificationManager
}

// NewMessageManager ...
func NewMessageManager(ctx *context.OrtooContext, client *model.Client, host string, notifyManager *NotificationManager) *MessageManager {

	return &MessageManager{
		seq:                 0,
		ctx:                 ctx,
		host:                host,
		client:              client,
		notificationManager: notifyManager,
	}
}

func (m *MessageManager) SetNotificationReceiver(receiver notificationReceiver) {
	if m.notificationManager != nil {
		m.notificationManager.SetReceiver(receiver)
	}
}

func (m *MessageManager) NextSeq() uint32 {
	currentSeq := m.seq
	m.seq++
	return currentSeq
}

// Connect ...
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

// Close ...
func (m *MessageManager) Close() error {
	if err := m.conn.Close(); err != nil {
		return errors.NewClientError(errors.ErrClientClose, err.Error())
	}
	if m.notificationManager != nil {
		m.notificationManager.Close()
	}
	return nil
}

func (m *MessageManager) Sync(pppList ...*model.PushPullPack) (*model.PushPullResponse, error) {
	request := model.NewPushPullRequest(m.NextSeq(), m.client, pppList...)
	m.ctx.Logger.Infof("send PUSHPULL REQUEST:%s", request.ToString())
	response, err := m.serviceClient.ProcessPushPull(m.ctx, request)
	if err != nil {
		return nil, log.OrtooErrorf(err, "fail to sync push pull")
	}
	m.ctx.Logger.Infof("receive PUSHPULL RESPONSE:%v", response.ToString())
	return response, nil
}

// ExchangeClientRequestResponse ...
func (m *MessageManager) ExchangeClientRequestResponse() error {
	request := model.NewClientRequest(m.NextSeq(), m.client)
	m.ctx.Logger.Infof("send CLIENT REQUEST:%s", request.ToString())
	response, err := m.serviceClient.ProcessClient(m.ctx, request)
	if err != nil {
		return log.OrtooErrorf(err, "fail to exchange clientRequestReply")
	}
	m.ctx.Logger.Infof("receive CLIENT RESPONSE: %s", response.ToString())
	return nil
}
