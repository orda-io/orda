package managers

import (
	"encoding/json"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/orda-io/orda/pkg/context"
	"github.com/orda-io/orda/pkg/errors"
	"github.com/orda-io/orda/pkg/model"
)

// NotifyManager manages notifications from Orda Server
type NotifyManager struct {
	client   mqtt.Client
	ctx      *context.ClientContext
	channel  chan *notificationMsg
	receiver notificationReceiver
}

type notificationReceiver interface {
	ReceiveNotification(topic string, notification model.Notification)
}

type pubSubNotificationType uint8

type notificationMsg struct {
	typeOf pubSubNotificationType
	topic  string
	msg    interface{}
}

const (
	notificationError pubSubNotificationType = iota
	notificationQuit
	notificationPushPull
)

// NewNotifyManager creates an instance of NotifyManager
func NewNotifyManager(ctx *context.ClientContext, pubSubAddr string, cm *model.Client) *NotifyManager {
	pubSubOpts := mqtt.NewClientOptions().
		AddBroker(pubSubAddr).
		SetClientID(cm.GetCUID()).
		SetUsername(cm.Alias)
	client := mqtt.NewClient(pubSubOpts)
	channel := make(chan *notificationMsg)
	return &NotifyManager{
		ctx:     ctx,
		client:  client,
		channel: channel,
	}
}

// SubscribeNotification subscribes notification for a topic.
func (its *NotifyManager) SubscribeNotification(topic string) errors.OrdaError {
	token := its.client.Subscribe(topic, 0, its.notificationSubscribeFunc)
	if token.Wait() && token.Error() != nil {
		return errors.ClientConnect.New(its.ctx.L(), "notification ", token.Error())
	}
	return nil
}

func (its *NotifyManager) notificationSubscribeFunc(client mqtt.Client, msg mqtt.Message) {
	notification := model.Notification{}
	if err := json.Unmarshal(msg.Payload(), &notification); err != nil {
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

// Connect make a connection with Orda notification server.
func (its *NotifyManager) Connect() errors.OrdaError {
	if token := its.client.Connect(); token.Wait() && token.Error() != nil {
		return errors.ClientConnect.New(its.ctx.L(), "notification server")
	}
	its.ctx.L().Infof("connect to notification server")
	go its.notificationLoop()
	return nil
}

// Close closes a connection with Orda notification server.
func (its *NotifyManager) Close() {
	its.channel <- &notificationMsg{
		typeOf: notificationQuit,
	}
	its.client.Disconnect(0)
}

// SetReceiver sets receiver which is going to receive notifications, i.e., DatatypeManager
func (its *NotifyManager) SetReceiver(receiver notificationReceiver) {
	its.receiver = receiver
}

func (its *NotifyManager) notificationLoop() {
	for {
		note := <-its.channel
		switch note.typeOf {
		case notificationError:
			err := note.msg.(error)
			its.ctx.L().Errorf("receive a notification error: %s", err.Error())
		case notificationQuit:
			its.ctx.L().Infof("quit notification loop")
			return
		case notificationPushPull:
			notification := note.msg.(model.Notification)
			its.receiver.ReceiveNotification(note.topic, notification)
		}
	}
}
