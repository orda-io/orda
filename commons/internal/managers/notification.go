package managers

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gogo/protobuf/proto"
	"github.com/knowhunger/ortoo/commons/context"
	"github.com/knowhunger/ortoo/commons/errors"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
)

type NotificationManager struct {
	client   mqtt.Client
	ctx      *context.OrtooContext
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

func NewNotificationManager(ctx *context.OrtooContext, pubSubAddr string) *NotificationManager {
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

func (n *NotificationManager) SubscribeNotification(topic string) error {
	if token := n.client.Subscribe(topic, 0, n.notificationSubscribeFunc); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (n *NotificationManager) notificationSubscribeFunc(client mqtt.Client, msg mqtt.Message) {
	notification := model.NotificationPushPull{}
	if err := proto.Unmarshal(msg.Payload(), &notification); err != nil {
		n.channel <- &notificationMsg{
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
	n.channel <- notificationPushPull
}

func (n *NotificationManager) Connect() error {
	if token := n.client.Connect(); token.Wait() && token.Error() != nil {
		return errors.NewClientError(errors.ErrClientConnect, "notification server")
	}
	n.ctx.Logger.Infof("connect to notification server")
	go n.notificationLoop()
	return nil
}

func (n *NotificationManager) Close() {
	n.client.Disconnect(0)
	n.channel <- &notificationMsg{
		typeOf: notificationQuit,
	}
}

func (n *NotificationManager) SetReceiver(receiver notificationReceiver) {
	n.receiver = receiver
}

func (n *NotificationManager) notificationLoop() {
	for {
		note := <-n.channel
		switch note.typeOf {
		case notificationError:
			err := note.msg.(error)
			_ = log.OrtooError(err)
		case notificationQuit:
			n.ctx.Logger.Infof("Quit notification loop")
			return
		case notificationPushPull:
			notification := note.msg.(model.NotificationPushPull)
			n.receiver.ReceiveNotification(note.topic, notification)
		}
	}
}
