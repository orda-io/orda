package client

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gogo/protobuf/proto"
	"github.com/knowhunger/ortoo/commons/context"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
	"google.golang.org/grpc"
)

// MessageManager is a manager exchanging request and response.
type MessageManager struct {
	seq            uint32
	host           string
	ctx            *context.OrtooContext
	client         *model.Client
	conn           *grpc.ClientConn
	serviceClient  model.OrtooServiceClient
	pubSubClient   mqtt.Client
	pubSubChan     chan *pubSubNotification
	pubSubReceiver notificationReceiver
}

type notificationReceiver interface {
	ReceiveNotification(topic string, notification model.NotificationPushPull)
}

type pubSubNotificationType uint8

const (
	pubSubError pubSubNotificationType = iota
	pubSubQuit
	pubSubNotificationPushPull
)

type pubSubNotification struct {
	typeOf pubSubNotificationType
	topic  string
	msg    interface{}
}

// NewMessageManager ...
func NewMessageManager(ctx *context.OrtooContext, client *model.Client, host string, pubSubAddr string) *MessageManager {
	pubSubOpts := mqtt.NewClientOptions().AddBroker(pubSubAddr)
	pubSubClient := mqtt.NewClient(pubSubOpts)
	pubSubChan := make(chan *pubSubNotification)
	return &MessageManager{
		seq:          0,
		ctx:          ctx,
		host:         host,
		client:       client,
		pubSubClient: pubSubClient,
		pubSubChan:   pubSubChan,
	}
}

func (m *MessageManager) SetNotificationReceiver(receiver notificationReceiver) {
	m.pubSubReceiver = receiver
}

// ExchangeClientRequestResponse ...
func (m *MessageManager) ExchangeClientRequestResponse() error {
	request := model.NewClientRequest(m.NextSeq(), m.client)
	response, err := m.serviceClient.ProcessClient(m.ctx, request)
	if err != nil {
		return log.OrtooErrorf(err, "fail to exchange clientRequestReply")
	}
	log.Logger.Infof("receive Client RESPONSE: %s", response.ToString())
	return nil

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
		return log.OrtooErrorf(err, "fail to connect to Ortoo Server")
	}
	m.conn = conn
	m.serviceClient = model.NewOrtooServiceClient(m.conn)

	if token := m.pubSubClient.Connect(); token.Wait() && token.Error() != nil {
		return log.OrtooErrorf(token.Error(), "fail to connect pub-sub")
	}

	go func() {
		for {
			pubSubNoti := <-m.pubSubChan
			switch pubSubNoti.typeOf {
			case pubSubError:
				err := pubSubNoti.msg.(error)
				_ = log.OrtooError(err)
			case pubSubQuit:
				log.Logger.Infof("Quit pubsub loop: %s", m.client.ToString())
				return
			case pubSubNotificationPushPull:
				notification := pubSubNoti.msg.(model.NotificationPushPull)
				m.pubSubReceiver.ReceiveNotification(pubSubNoti.topic, notification)
			}
		}
	}()
	return nil
}

// Close ...
func (m *MessageManager) Close() error {
	if err := m.conn.Close(); err != nil {
		return log.OrtooErrorf(err, "fail to close grpc connection")
	}
	m.pubSubClient.Disconnect(0)
	m.pubSubChan <- &pubSubNotification{
		typeOf: pubSubQuit,
		msg:    nil,
	}
	return nil
}

func (m *MessageManager) Sync(pppList ...*model.PushPullPack) (*model.PushPullResponse, error) {

	request := model.NewPushPullRequest(m.NextSeq(), m.client, pppList...)
	pushPullResponse, err := m.serviceClient.ProcessPushPull(m.ctx, request)
	if err != nil {
		return nil, log.OrtooErrorf(err, "fail to sync push pull")
	}
	log.Logger.Infof("receive PushPull RESPONSE:%v", pushPullResponse.ToString())
	return pushPullResponse, nil
}

func (m *MessageManager) SubscribePubSub(topic string) error {
	if token := m.pubSubClient.Subscribe(topic, 0, m.pubSubSubscribeFunc); token.Wait() && token.Error() != nil {
		return token.Error()
		// icImpl.HandleError(errors.NewDatatypeError(errors.ErrDatatypeSubscribe, token.Error()))
	}
	return nil
}

func (m *MessageManager) pubSubSubscribeFunc(client mqtt.Client, msg mqtt.Message) {
	notification := model.NotificationPushPull{}
	if err := proto.Unmarshal(msg.Payload(), &notification); err != nil {
		m.pubSubChan <- &pubSubNotification{
			typeOf: pubSubError,
			msg:    err,
		}
		return
	}

	notificationPushPull := &pubSubNotification{
		typeOf: pubSubNotificationPushPull,
		topic:  msg.Topic(),
		msg:    notification,
	}
	m.pubSubChan <- notificationPushPull
}
