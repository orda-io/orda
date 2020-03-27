package main

import (
	"fmt"
	"github.com/knowhunger/ortoo/ortoo"
	"github.com/knowhunger/ortoo/ortoo/model"
)

func main() {

	clientConf := &ortoo.ClientConfig{
		ServerAddr:       "localhost:19061",         // Ortoo Server address.
		NotificationAddr: "localhost:1883",          // notification server address.
		CollectionName:   "hello_world",             // the collection name of MongoDB which the client participates in.
		SyncType:         model.SyncType_NOTIFIABLE, // syncType that is notified in real-time from notification server.
	}

	client1 := ortoo.NewClient(clientConf, "client1") // create a client with alias "client1".
	if err := client1.Connect(); err != nil {         // connect to Ortoo server
		_ = fmt.Errorf("fail client to connect an Ortoo server: %v", err.Error())
		return
	}
	defer func() {
		if err := client1.Close(); err != nil { // close the client
			_ = fmt.Errorf("fail to close client: %v", err.Error())
		}
	}()

	// intCounter := client1.CreateDatatype()
	//
	// intCounter.Increase()
	// intCounter.IncreaseBy(3)
	// intCounter.DoTransaction(func() {
	// 	intCounter.Increase()
	// 	intCounter.Decrease()
	// })
	// c.Sync()
	// intCounter.DoTransaction(fun)

}
