package main

import (
	"fmt"
	"github.com/orda-io/orda/client/pkg/errors"
	"github.com/orda-io/orda/client/pkg/model"
	"github.com/orda-io/orda/client/pkg/orda"
	"sync"
)

func main() {
	conf := &orda.ClientConfig{
		ServerAddr:       "localhost:19061",
		NotificationAddr: "tcp://localhost:18181",
		CollectionName:   "hello_world",
		SyncType:         model.SyncType_REALTIME,
	}

	client1 := orda.NewClient(conf, "client1")
	client2 := orda.NewClient(conf, "client2")

	if err := client1.Connect(); err != nil {
		panic("fail to connect client1 to an Orda server:" + err.Error())
	}
	if err := client2.Connect(); err != nil {
		panic("fail to connect client2 to an Orda server" + err.Error())
	}
	defer func() {
		if err := client1.Close(); err != nil {
			_ = fmt.Errorf("fail to close client1: %v", err.Error())
		}
		if err := client2.Close(); err != nil {
			_ = fmt.Errorf("fail to close client2: %v", err.Error())
		}
	}()
	wg := &sync.WaitGroup{}
	client1.CreateDocument("sampleDoc", orda.NewHandlers(
		func(dt orda.Datatype, old model.StateOfDatatype, new model.StateOfDatatype) {
			if new == model.StateOfDatatype_SUBSCRIBED {
				wg.Done()
			}
		},
		func(dt orda.Datatype, opList []interface{}) {

		},
		func(dt orda.Datatype, errs ...errors.OrdaError) {

		},
	))

}
