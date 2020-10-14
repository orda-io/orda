package managers

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gogo/protobuf/proto"
	"github.com/knowhunger/ortoo/pkg/context"
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/model"
)

// NotificationManager manages notifications from Ortoo Server
type NotificationManager struct {
	client   mqtt.Client
	ctx      context.OrtooContext
	channel  chan *notificationMsg
	receiver notificationReceiver
}

type notificationReceiver interface {
	ReceiveNotification(topic string, notification model.NotificationPushPull)
}

type pubSubNotificationType uint8

const (
	notificationError pubSubNotificationType = iota
	notificationQuit
	notificationPushPull
)

// NewNotificationManager creates an instance of NotificationManager
func NewNotificationManager(ctx context.OrtooContext, pubSubAddr string) *NotificationManager {
	pubSubOpts := mqtt.NewClientOptions().AddBroker(pubSubAddr)
	client := mqtt.NewClient(pubSubOpts)
	channel := make(chan *notificationMsg)
	return &NotificationManager{
		ctx:     ctx,
		client:  client,
		channel: channel,
	}
}

type notificationMsg struct {
	typeOf pubSubNotificationType
	topic  string
	msg    interface{}
}

// SubscribeNotification subscribes notification for a topic.
func (its *NotificationManager) SubscribeNotification(topic string) errors.OrtooError {
	token := its.client.Subscribe(topic, 0, its.notificationSubscribeFunc)
	if token.Wait() && token.Error() != nil {
		return errors.ClientConnect.New(its.ctx.L(), "notification ", token.Error())
	}
	return nil
}

func (its *NotificationManager) notificationSubscribeFunc(client mqtt.Client, msg mqtt.Message) {
	notification := model.NotificationPushPull{}
	if err := proto.Unmarshal(msg.Payload(), &notification); err != nil {
		its.channel <- &notificationMsg{
			typeOf: notificationError,
			msg:    err,
		}
		return
	}

	notificationPushPull := &notificationMsg{
		typeOf: notificationPushPull,
		topic:  msg.Topic(),
		msg:    notification,
	}
	its.channel <- notificationPushPull
}

// Connect make a connection with Ortoo notification server.
func (its *NotificationManager) Connect() errors.OrtooError {
	if token := its.client.Connect(); token.Wait() && token.Error() != nil {
		return errors.ClientConnect.New(its.ctx.L(), "notification server")
	}
	its.ctx.L().Infof("connect to notification server")
	go its.notificationLoop()
	return nil
}

// Close closes a connection with Ortoo notification server.
func (its *NotificationManager) Close() {
	its.channel <- &notificationMsg{
		typeOf: notificationQuit,
	}
	its.client.Disconnect(0)
}

// SetReceiver sets receiver which is going to receive notifications, i.e., DatatypeManager
func (its *NotificationManager) SetReceiver(receiver notificationReceiver) {
	its.receiver = receiver
}

func (its *NotificationManager) notificationLoop() {
	for {
		note := <-its.channel
		switch note.typeOf {
		case notificationError:
			err := note.msg.(error)
			its.ctx.L().Errorf("receive a notification error: %s", err.Error())
			break
		case notificationQuit:
			its.ctx.L().Infof("Quit notification loop")
			return
		case notificationPushPull:
			notification := note.msg.(model.NotificationPushPull)
			its.receiver.ReceiveNotification(note.topic, notification)
			break
		}
	}
}
