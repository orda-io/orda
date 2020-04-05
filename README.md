# Ortoo
[Ortoo (or Yam)](https://en.wikipedia.org/wiki/Yam_(route) was a mongolian messenger system. Ortoo project is a multi-device data synchronization platform based on MongoDB (which could be other document databases). Ortoo is based on CRDT(conflict-free data types), which enables operation-based syncronization.  


## Getting started

### Working envirnment (Maybe work on less versions of them)
 - go 1.13.5
 - docker 18.09.2 (for running MongoDB)
 - [MongoDB latest](https://hub.docker.com/_/mongo)
 - MQTT: [eclipse mosquitto latest](https://hub.docker.com/_/eclipse-mosquitto) 
 - docker-compose
 - gogo/protobuf (how to install: http://google.github.io/proto-lens/installing-protoc.html)
 
### Install
 ```bash
 # git clone https://github.com/knowhunger/ortoo.git
 # cd ortoo 
 # make docker-up
 # make protoc-gen
 ```

## How to use Ortoo

### Run Ortoo Server
```go

```

### Make a client
 - An Ortoo client manages the connection with the Ortoo server and synchronization of Ortoo datatypes.   
 - An Ortoo client participates in a collection of MongoDB, which means that the snapshot of any created datatype is written to the collection of MongoDB.   
```go
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
```
## Use datatypes


### IntCounter
### HashMap
### List
### Transaction
### Inside MongoDB 


