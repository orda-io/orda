package notification

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gogo/protobuf/proto"
	"github.com/knowhunger/ortoo/pkg/context"
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/server/mongodb/schema"
)

// Notifier is a struct that takes responsibility for notification
type Notifier struct {
	mqttClient mqtt.Client
}

// NewNotifier creates an instance of Notifier
func NewNotifier(ctx context.OrtooContext, pubSubAddr string) (*Notifier, errors.OrtooError) {
	Opts := mqtt.NewClientOptions().AddBroker(pubSubAddr)
	// Opts.SetClientID("ortoo-server")

	client := mqtt.NewClient(Opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, errors.ServerInit.New(ctx.L(), token.Error())
	}
	return &Notifier{mqttClient: client}, nil
}

// NotifyAfterPushPull enables server to send a notification to MQTT server
func (n *Notifier) NotifyAfterPushPull(
	ctx context.OrtooContext,
	collectionName string,
	client *schema.ClientDoc,
	datatype *schema.DatatypeDoc,
	sseq uint64,
) errors.OrtooError {
	topic := fmt.Sprintf("%s/%s", collectionName, datatype.Key)
	msg := model.NotificationPushPull{
		CUID: client.CUID,
		DUID: datatype.DUID,
		Sseq: sseq,
	}
	bMsg, err := proto.Marshal(&msg)
	if err != nil {
		return errors.ServerNotify.New(ctx.L(), err.Error())
	}
	if token := n.mqttClient.Publish(topic, 0, false, bMsg); token.Wait() && token.Error() != nil {
		return errors.ServerNotify.New(ctx.L(), token.Error())
	}
	ctx.L().Infof("notify datatype topic:(%s) with sseq:%d by %s", topic, sseq, client.GetClientSummary())
	return nil
}
