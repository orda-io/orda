package main

import (
	"fmt"
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/pkg/ortoo"
	"sync"
)

const (
	intCounterKey = "intCounter_example"
)

func main() {
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		client1, err := createClient("client1")
		if err != nil {
			panic(err)
		}
		defer closeClient(client1)
		createIntCounter(client1)
		wg.Done()
	}()
	go func() {
		client2, err := createClient("client2")
		if err != nil {
			panic(err)
		}
		defer closeClient(client2)
		createOrSubscribeIntCounter(client2)
		wg.Done()
	}()
	wg.Wait()
	client3, err := createClient("client3")
	if err != nil {
		panic(err)
	}
	defer closeClient(client3)
	subscribeIntCounter(client3)
}

func createClient(alias string) (ortoo.Client, error) {
	clientConf := &ortoo.ClientConfig{
		ServerAddr:       "localhost:19061",         // Ortoo Server address.
		NotificationAddr: "localhost:11883",         // notification server address.
		CollectionName:   "hello_world",             // the collection name of MongoDB which the client participates in.
		SyncType:         model.SyncType_NOTIFIABLE, // syncType that is notified in real-time from notification server.
	}

	client1 := ortoo.NewClient(clientConf, alias) // create a client with alias "client1".
	if err := client1.Connect(); err != nil {     // connect to Ortoo server
		_ = fmt.Errorf("fail client to connect an Ortoo server: %v", err.Error())
		return nil, err
	}
	return client1, nil
}

func closeClient(client ortoo.Client) {
	if err := client.Close(); err != nil { // close the client
		_ = fmt.Errorf("fail to close client: %v", err.Error())
	}
}

func createIntCounter(client ortoo.Client) {
	intCounter := client.CreateCounter(intCounterKey, ortoo.NewHandlers(
		func(dt ortoo.Datatype, old model.StateOfDatatype, new model.StateOfDatatype) {
			fmt.Printf("Can see how to change the state of datatype: %s => %s\n", old.String(), new.String())
		},
		func(dt ortoo.Datatype, opList []interface{}) {
			fmt.Printf("Received remote operations\n")
			for op := range opList {
				fmt.Printf("%v", op)
			}
		},
		func(dt ortoo.Datatype, errs ...errors.OrtooError) {
			fmt.Printf("Can handle error: %v", errs)
		}))
	if err1 := client.Sync(); err1 != nil {
		panic(err1)
	}
	val, err2 := intCounter.IncreaseBy(5)
	if err2 != nil {
		panic(err2)
	}
	fmt.Printf("After increase: %d\n", val)

	if err3 := client.Sync(); err3 != nil {
		panic(err3)
	}
}

func createOrSubscribeIntCounter(client ortoo.Client) {
	intCounter := client.SubscribeOrCreateCounter(intCounterKey, ortoo.NewHandlers(
		func(dt ortoo.Datatype, old model.StateOfDatatype, new model.StateOfDatatype) {
			fmt.Printf("Can see how to change the state of datatype: %s => %s\n", old.String(), new.String())
		},
		func(dt ortoo.Datatype, opList []interface{}) {
			fmt.Printf("Received remote operations\n")
			for op := range opList {
				fmt.Printf("%v", op)
			}
		},
		func(dt ortoo.Datatype, errs ...errors.OrtooError) {
			fmt.Printf("Can handle error: %v", errs)
		}))
	if err := client.Sync(); err != nil {
		panic(err)
	}
	val, err := intCounter.IncreaseBy(5)
	if err != nil {
		panic(err)
	}
	fmt.Printf("After increase: %d\n", val)

	if err2 := client.Sync(); err2 != nil {
		panic(err2)
	}
}

func subscribeIntCounter(client ortoo.Client) {
	intCounter := client.SubscribeCounter(intCounterKey, ortoo.NewHandlers(
		func(dt ortoo.Datatype, old model.StateOfDatatype, new model.StateOfDatatype) {
			fmt.Printf("Can see how to change the state of datatype: %s => %s\n", old.String(), new.String())
		},
		func(dt ortoo.Datatype, opList []interface{}) {
			fmt.Printf("Received remote operations\n")
			for op := range opList {
				fmt.Printf("%v", op)
			}
		},
		func(dt ortoo.Datatype, errs ...errors.OrtooError) {
			fmt.Printf("Can handle error: %v", errs)
		}))
	if err1 := client.Sync(); err1 != nil {
		panic(err1)
	}
	val, err := intCounter.IncreaseBy(5)
	if err != nil {
		panic(err)
	}
	fmt.Printf("After increase: %d\n", val)

	if err2 := client.Sync(); err2 != nil {
		panic(err2)
	}
}
