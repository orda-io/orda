package notification

import (
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/orda-io/orda/pkg/context"
	"github.com/orda-io/orda/pkg/errors"
	"github.com/orda-io/orda/pkg/model"
	"github.com/orda-io/orda/server/mongodb/schema"
)

// Notifier is a struct that takes responsibility for notification
type Notifier struct {
	mqttClient mqtt.Client
}

// NewNotifier creates an instance of Notifier
func NewNotifier(ctx context.OrdaContext, pubSubAddr string, serverName string) (*Notifier, errors.OrdaError) {
	opts := mqtt.NewClientOptions().AddBroker(pubSubAddr).SetUsername(serverName)
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, errors.ServerInit.New(ctx.L(), token.Error())
	}
	return &Notifier{mqttClient: client}, nil
}

// NotifyAfterPushPull enables server to send a notification to MQTT server
func (n *Notifier) NotifyAfterPushPull(
	ctx context.OrdaContext,
	collectionName string,
	client *schema.ClientDoc,
	datatype *schema.DatatypeDoc,
	sseq uint64,
) errors.OrdaError {
	topic := fmt.Sprintf("%s/%s", collectionName, datatype.Key)
	msg := model.Notification{
		CUID: client.CUID,
		DUID: datatype.DUID,
		Sseq: sseq,
	}
	bMsg, err := json.Marshal(&msg)
	ctx.L().Infof("%s", bMsg)
	if err != nil {
		return errors.ServerNotify.New(ctx.L(), err.Error())
	}
	if token := n.mqttClient.Publish(topic, 0, false, bMsg); token.Wait() && token.Error() != nil {
		return errors.ServerNotify.New(ctx.L(), token.Error())
	}
	ctx.L().Infof("notify datatype topic:(%s) with sseq:%d by %s", topic, sseq, client.ToString())
	return nil
}
