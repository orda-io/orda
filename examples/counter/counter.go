package main

import (
	"fmt"
	"sync"

	"github.com/orda-io/orda/pkg/errors"
	"github.com/orda-io/orda/pkg/model"
	"github.com/orda-io/orda/pkg/orda"
)

const (
	counterKey = "intCounter_example"
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
		createCounter(client1)
		wg.Done()
	}()
	go func() {
		client2, err := createClient("client2")
		if err != nil {
			panic(err)
		}
		defer closeClient(client2)
		createOrSubscribeCounter(client2)
		wg.Done()
	}()
	wg.Wait()
	client3, err := createClient("client3")
	if err != nil {
		panic(err)
	}
	defer closeClient(client3)
	subscribeCounter(client3)
}

func createClient(alias string) (orda.Client, error) {
	clientConf := &orda.ClientConfig{
		ServerAddr:       "localhost:29065",       // Orda Server address. The port 29065 is the port when it is running on docker
		NotificationAddr: "localhost:11883",       // notification server address.
		CollectionName:   "hello_world",           // the collection name of MongoDB which the client participates in.
		SyncType:         model.SyncType_REALTIME, // syncType that is notified in real-time from notification server.
	}

	client1 := orda.NewClient(clientConf, alias) // create a client with alias "client1".
	if err := client1.Connect(); err != nil {    // connect to Orda server
		_ = fmt.Errorf("fail client to connect an Orda server: %v", err.Error())
		return nil, err
	}
	return client1, nil
}

func closeClient(client orda.Client) {
	if err := client.Close(); err != nil { // close the client
		_ = fmt.Errorf("fail to close client: %v", err.Error())
	}
}

func createCounter(client orda.Client) {
	intCounter := client.CreateCounter(counterKey, orda.NewHandlers(
		func(dt orda.Datatype, old model.StateOfDatatype, new model.StateOfDatatype) {
			fmt.Printf("Can see how to change the state of datatype: %s => %s\n", old.String(), new.String())
		},
		func(dt orda.Datatype, opList []interface{}) {
			fmt.Printf("Received remote operations\n")
			for op := range opList {
				fmt.Printf("%v", op)
			}
		},
		func(dt orda.Datatype, errs ...errors.OrdaError) {
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

func createOrSubscribeCounter(client orda.Client) {
	intCounter := client.SubscribeOrCreateCounter(counterKey, orda.NewHandlers(
		func(dt orda.Datatype, old model.StateOfDatatype, new model.StateOfDatatype) {
			fmt.Printf("Can see how to change the state of datatype: %s => %s\n", old.String(), new.String())
		},
		func(dt orda.Datatype, opList []interface{}) {
			fmt.Printf("Received remote operations\n")
			for op := range opList {
				fmt.Printf("%v", op)
			}
		},
		func(dt orda.Datatype, errs ...errors.OrdaError) {
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

func subscribeCounter(client orda.Client) {
	intCounter := client.SubscribeCounter(counterKey, orda.NewHandlers(
		func(dt orda.Datatype, old model.StateOfDatatype, new model.StateOfDatatype) {
			fmt.Printf("Can see how to change the state of datatype: %s => %s\n", old.String(), new.String())
		},
		func(dt orda.Datatype, opList []interface{}) {
			fmt.Printf("Received remote operations\n")
			for op := range opList {
				fmt.Printf("%v", op)
			}
		},
		func(dt orda.Datatype, errs ...errors.OrdaError) {
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
