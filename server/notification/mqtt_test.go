package notification_test

import (
	"fmt"
	"sync"
	"testing"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func TestMqttPubSub(t *testing.T) {

	const TOPIC = "mytopic/test"

	opts := mqtt.NewClientOptions().AddBroker("tcp://127.0.0.1:18181")
	opts.SetClientID("client1")
	client1 := mqtt.NewClient(opts)
	if token := client1.Connect(); token.Wait() && token.Error() != nil {
		t.Fatal(token.Error())
	}
	opts.SetClientID("client2")
	client2 := mqtt.NewClient(opts)
	if token2 := client2.Connect(); token2.Wait() && token2.Error() != nil {
		t.Fatal(token2.Error())
	}

	var wg sync.WaitGroup
	wg.Add(2)

	var subscribeFn = func(client mqtt.Client, msg mqtt.Message) {
		reader := client.OptionsReader()
		fmt.Printf("at %v: %s\n", reader.ClientID(), string(msg.Payload()))
		wg.Done()
	}

	if token := client1.Subscribe(TOPIC, 0, subscribeFn); token.Wait() && token.Error() != nil {
		t.Fatal(token.Error())
	}

	if token2 := client2.Subscribe(TOPIC, 0, subscribeFn); token2.Wait() && token2.Error() != nil {
		t.Fatal(token2.Error())
	}

	if token := client1.Publish(TOPIC, 0, false, "mymessage"); token.Wait() && token.Error() != nil {
		t.Fatal(token.Error())
	}
	wg.Wait()
}
