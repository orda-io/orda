package notification

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gogo/protobuf/proto"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/ortoo/model"
	"github.com/knowhunger/ortoo/server/mongodb/schema"
)

// Notifier is a struct that takes responsibility for notification
type Notifier struct {
	pubSubClient mqtt.Client
}

// NewNotifier creates an instance of Notifier
func NewNotifier(pubSubAddr string) (*Notifier, error) {
	pubSubOpts := mqtt.NewClientOptions().AddBroker(pubSubAddr)
	pubSubClient := mqtt.NewClient(pubSubOpts)
	if token := pubSubClient.Connect(); token.Wait() && token.Error() != nil {
		return nil, log.OrtooError(token.Error())
	}
	return &Notifier{pubSubClient: pubSubClient}, nil
}

// NotifyAfterPushPull enables server to send a notification to MQTT server
func (n *Notifier) NotifyAfterPushPull(collectionName string, client *schema.ClientDoc, datatype *schema.DatatypeDoc, sseq uint64) error {
	topic := fmt.Sprintf("%s/%s", collectionName, datatype.Key)
	msg := model.NotificationPushPull{
		CUID: client.CUID,
		DUID: datatype.DUID,
		Sseq: sseq,
	}
	bMsg, err := proto.Marshal(&msg)
	if err != nil {
		return log.OrtooError(err)
	}
	if token := n.pubSubClient.Publish(topic, 0, false, bMsg); token.Wait() && token.Error() != nil {
		return log.OrtooError(token.Error())
	}
	log.Logger.Infof("notify %s with sseq:%d by %s", datatype, sseq, client)
	return nil
}
