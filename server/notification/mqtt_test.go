package notification_test

import (
	"fmt"
	"sync"
	"testing"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func TestMqttPubSub(t *testing.T) {

	const TOPIC = "mytopic/test"

	opts := mqtt.NewClientOptions().AddBroker("tcp://localhost:1883")

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		t.Fatal(token.Error())
	}

	client2 := mqtt.NewClient(opts)
	if token2 := client2.Connect(); token2.Wait() && token2.Error() != nil {
		t.Fatal(token2.Error())
	}

	var wg sync.WaitGroup
	wg.Add(2)

	if token := client.Subscribe(TOPIC, 0, func(client mqtt.Client, msg mqtt.Message) {
		if string(msg.Payload()) != "mymessage" {
			t.Fatalf("want mymessage, got %s", msg.Payload())
		}
		wg.Done()
	}); token.Wait() && token.Error() != nil {
		t.Fatal(token.Error())
	}

	if token2 := client2.Subscribe(TOPIC, 0, func(client2 mqtt.Client, msg mqtt.Message) {
		fmt.Printf("at client2: %s\n", string(msg.Payload()))
		wg.Done()
	}); token2.Wait() && token2.Error() != nil {
		t.Fatal(token2.Error())
	}

	if token := client.Publish(TOPIC, 0, false, "mymessage"); token.Wait() && token.Error() != nil {
		t.Fatal(token.Error())
	}
	wg.Wait()
}
