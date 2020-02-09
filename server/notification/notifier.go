package notification

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gogo/protobuf/proto"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
)

type Notifier struct {
	pubSubClient mqtt.Client
}

func NewNotifier(pubSubAddr string) (*Notifier, error) {
	pubSubOpts := mqtt.NewClientOptions().AddBroker(pubSubAddr)
	pubSubClient := mqtt.NewClient(pubSubOpts)
	if token := pubSubClient.Connect(); token.Wait() && token.Error() != nil {
		return nil, log.OrtooError(token.Error())
	}
	return &Notifier{pubSubClient: pubSubClient}, nil
}

func (n *Notifier) NotifyAfterPushPull(collectionName, key, cuid, duid string, sseq uint64) error {
	topic := fmt.Sprintf("%s/%s", collectionName, key)
	msg := model.NotificationPushPull{
		CUID: cuid,
		DUID: duid,
		Sseq: sseq,
	}
	bMsg, err := proto.Marshal(&msg)
	if err != nil {
		return log.OrtooError(err)
	}
	if token := n.pubSubClient.Publish(topic, 0, false, bMsg); token.Wait() && token.Error() != nil {
		return log.OrtooError(token.Error())
	}
	return nil
}
